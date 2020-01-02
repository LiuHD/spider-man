package factory

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"martin/spider_man/util"
	"math/rand"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Resource struct {
	Uri      string
	TryTimes int
	Error    error
	Refer    string
	Id       string
	Num      int
}

type Seeder struct {
	Resource
}

type Pleasure struct {
	Resource

	Entity []byte
	Name   string
	Ext    string
}

const STORE_ROOT = "data1"

type Worker struct {
	Id           string
	SeederChan   chan Seeder
	PleasureChan chan []Pleasure
	Dispatcher   *Dispatcher
}

func (w *Worker) Run() {
	w.Work()
}

func (w *Worker) Work() {
	for {
		select {
		case s := <-w.SeederChan:
			text, err := w.Dig(s.Resource)
			if err != nil {
				log.Fatalln("旷工"+w.Id+"来报，挖取出错了", err)
				go func() { w.SeederChan <- s }()
			}
			if len(text) == 0 {
				go func() { w.SeederChan <- s }()
				continue
			}
			if len(s.Id) > 0 && s.Num == 1 {
				w.PreSave(s.Id, text)
			}

			seeders, pleasures := w.Pick(s.Uri, s.Id, s.Num, text)
			if s.Uri != LIST_URI && len(seeders) == 0 && len(pleasures) == 0 {
				log.Fatalln("有个哑炮，检查一下", s.Uri, s.Num, s.Id)
			}
			if len(seeders) > 0 {
				go func() {
					for _, seeder := range seeders {
						w.SeederChan <- seeder
					}
				}()
			}
			if len(pleasures) > 0 {
				go func() { w.PleasureChan <- pleasures }()
			}
		case ps := <-w.PleasureChan:
			errPleasure := []Pleasure{}
			for _, p := range ps {
				entity, err := w.Dig(p.Resource)
				if err != nil {
					log.Println("旷工"+w.Id+"来报，搬运出错了，", err)
					errPleasure = append(errPleasure, p)
					continue
				}
				if len(entity) == 0 {
					continue
				}
				log.Println("要存图片了", p.Id, p.Name, p.Ext)
				p.Name = strings.Trim(p.Name, "/")
				err = util.SaveToFile(path.Join(STORE_ROOT, p.Id, strings.Replace(p.Name, "/", "_", -1)+"."+p.Ext), entity)
				if err != nil {
					log.Println("旷工"+w.Id+"来报，储存出错了，", err)
					errPleasure = append(errPleasure, p)
				}
				w.Dispatcher.SetDone(p.Num, p.Id)
			}
			if len(errPleasure) > 0 {
				go func() { w.PleasureChan <- errPleasure }()
			}
		}
	}
}

func (w *Worker) Pick(oriUri string, Id string, num int, text []byte) (seeders []Seeder, pleasures []Pleasure) {
	if num == 1 && oriUri != LIST_URI {
		if !w.Dispatcher.Exist(oriUri) {
			totalNumReg := regexp.MustCompile(`'dots'>…</span><a href='https://www\.mzitu\.com/` + Id + `/(\d{1,3})'`)
			tmp := totalNumReg.FindSubmatch(text)
			if len(tmp) == 0 {
				log.Fatalln(Id, oriUri+"没有找到最后一页")
			}
			totalNum, err := strconv.Atoi(string(tmp[1]))
			if err != nil {
				log.Fatalln(err)
			}
			w.Dispatcher.Add(Id, oriUri, totalNum, 0)
		}
	}
	if oriUri == LIST_URI {
		seederReg := regexp.MustCompile(`https://www\.mzitu\.com/(\d{1,6})(/(\d{1,3}))?`)
		seederUris := seederReg.FindAllSubmatch(text, -1)
		for _, seederUri := range seederUris {
			if !w.Dispatcher.Exist(string(seederUri[1])) {
				log.Println("拣到seeder", string(seederUri[1]))
				seeders = append(seeders, Seeder{
					Resource{Uri: string(seederUri[0]), Id: string(seederUri[1]), Num: 1}})

			}
		}
	}
	pleasureReg := regexp.MustCompile(`https://i\d{0,2}\.(?:meizitu\.net|mmzztt\.com)/(\d.*?)\.(jpg|jpeg|png|bmp|gif)`)
	pleasureUris := pleasureReg.FindAllSubmatch(text, -1)
	for _, pleasureUri := range pleasureUris {
		log.Println("拣到pleasure", string(pleasureUri[0]))
		pleasures = append(pleasures, Pleasure{
			Name:     strconv.Itoa(num) + "_" + string(pleasureUri[1]),
			Ext:      string(pleasureUri[2]),
			Resource: Resource{Uri: string(pleasureUri[0]), Refer: oriUri, Id: Id, Num: num}})
	}

	return
}

func (w *Worker) Dig(res Resource) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, res.Uri, new(bytes.Buffer))
	if err != nil {
		log.Fatalln(err)
	}
	addHeader(req)
	if len(res.Refer) > 0 {
		req.Header.Add("referer", res.Refer)
	}
	log.Println("旷工炸了一下", res.Uri)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res.TryTimes++
		res.Error = err
		return []byte{}, err
	}
	var str []byte
	defer resp.Body.Close()
	str, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	if encode, ok := resp.Header["Content-Encoding"]; ok {
		tmp := strings.Join(encode, "")
		if strings.Contains(tmp, "gzip") {
			str, err = util.DecodeGzip(str)
			resp.Header.Del("Content-Encoding")
		}
		if strings.Contains(tmp, "deflate") {
			str, err = util.DecodeDeflate(str)
			resp.Header.Del("Content-Encoding")
		}
	}

	//time.Sleep(time.Microsecond * time.Duration(1000+rand.Intn(800)))

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			log.Println("挖得太快了，有塌陷的危险，停一停")
			time.Sleep(time.Second * time.Duration(rand.Intn(8)))
			return []byte{}, nil
		}
		res.TryTimes++
		err = fmt.Errorf("请求返回错误 %d %s", resp.StatusCode, str)
		res.Error = err
		return []byte{}, err
	}
	return str, nil
}

func (w *Worker) PreSave(Id string, textByte []byte) {
	titleReg := regexp.MustCompile(`<h2 class="main-title">(.+?)</h2>`)
	tagReg := regexp.MustCompile(`rel="tag">(.*?)</a>`)
	dateReg := regexp.MustCompile(`<span>发布于\s(.*?)</span>`)
	categoryReg := regexp.MustCompile(`rel="category tag">(.*?)</a></span>`)
	text := string(textByte)
	if len(titleReg.FindStringSubmatch(text)) == 0 {
		log.Fatalln(Id, text)
	}
	title := titleReg.FindStringSubmatch(text)[1]
	tagsRaw := tagReg.FindAllStringSubmatch(text, -1)
	date := dateReg.FindStringSubmatch(text)[1]
	category := categoryReg.FindStringSubmatch(text)[1]

	tags := []string{}
	for _, t := range tagsRaw {
		tags = append(tags, t[1])
	}
	err := util.SaveToFile(path.Join(STORE_ROOT, Id, "info.txt"),
		[]byte(fmt.Sprintf("title:%s\ntags:%s\ncategory:%s\ndate:%s\n", title, strings.Join(tags, ","), category, date)))
	if err != nil {
		log.Fatalln("旷工"+w.Id+"来报, 建立仓库失败", err)
	}
}

func addHeader(req *http.Request) {
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("accept-encoding", "gzip")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7,ja;q=0.6,zh-HK;q=0.5")
	req.Header.Add("upgrade-insecure-requests", "0")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
}
