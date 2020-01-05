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

type Worker struct {
	Id           string
	SeederChan   chan Seeder
	PleasureChan chan []Pleasure
	Dispatcher   *Dispatcher
}

func (w *Worker) Run() {
	for {
		select {
		case s := <-w.SeederChan:
			text, err := w.Dig(s.Resource)
			if err != nil {
				log.Fatalln("旷工"+w.Id+"来报，挖取出错了", err)
			}
			if len(text) == 0 {
				log.Println("抓到空了", s.Uri)
				go func() { w.SeederChan <- s }()
				continue
			}
			seeders, pleasures := w.Pick(s, text)
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
					log.Println("抓到空了", p.Uri)
					continue
				}
				log.Println("要存图片了", p.Id, p.Name, p.Ext)
				p.Name = strings.Trim(p.Name, "/")
				err = util.SaveToFile(path.Join("data", w.Dispatcher.Ctx.SiteName, p.Id, strings.Replace(p.Name, "/", "_", -1)+"."+p.Ext), entity)
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

func (w *Worker) Pick(s Seeder, text []byte) (seeders []Seeder, pleasures []Pleasure) {
	if s.Num == 1 && len(s.Id) > 0 {
		if !w.Dispatcher.Exist(s.Id) {
			w.PreSave(s.Uri, s.Id, text)
			//添加子页面，如果可以计算出来的话
			totalNumReg := regexp.MustCompile(w.Dispatcher.Config.TotalPageReg)
			tmp := totalNumReg.FindStringSubmatch(string(text))
			totalNum, _ := strconv.Atoi(string(tmp[1]))
			pageUrlGenerator := w.Dispatcher.Config.PageUrlGenerator
			if pageUrlGenerator != nil {
				num := 0
				for num < totalNum {
					num++
					seeders = append(seeders, Seeder{Resource{Uri: pageUrlGenerator(s.Uri, s.Id, num)}})
				}
			}
		}
	}
	if len(s.Id) == 0 {
		seederReg := regexp.MustCompile(w.Dispatcher.Config.SeederReg)
		seederUris := seederReg.FindAllSubmatch(text, -1)
		if len(seederUris) == 0 {
			log.Fatalln("没找到seeder，关注一下", string(text))
		}
		for _, seederUri := range seederUris {
			if !w.Dispatcher.Exist(string(seederUri[1])) {
				log.Println("拣到seeder", string(seederUri[1]))
				seeders = append(seeders, Seeder{
					Resource{Uri: string(seederUri[0]), Id: string(seederUri[1]), Num: 1}})

			}
		}
		seeders = UniqueSeeder(seeders)
	} else if len(w.Dispatcher.Config.NextPageReg) > 0 {
		seederReg := regexp.MustCompile(w.Dispatcher.Config.NextPageReg)
		nextPageUri := seederReg.FindSubmatch(text)
		//todo: 逻辑缺失
		if nextPageUri == nil {
			log.Fatalln("没有下一页")
		}
		page, _ := strconv.Atoi(string(nextPageUri[3]))
		if !w.Dispatcher.PleasureExist(string(nextPageUri[2]), page) {
			log.Println("拣到seeder", string(nextPageUri[1]))
			w.Dispatcher.PleasureAdd(s.Uri, string(nextPageUri[2]), string(nextPageUri[1]), page)
			seeders = append(seeders, Seeder{
				Resource{Uri: string(nextPageUri[1]), Id: string(nextPageUri[2]), Num: page}})
		}
	}
	if len(s.Id) > 0 {
		pleasureReg := regexp.MustCompile(w.Dispatcher.Config.PleasureReg)
		pleasureUris := pleasureReg.FindAllSubmatch(text, -1)
		if len(pleasureUris) == 0 {
			log.Fatalln("没找到pleasure，关注一下", string(text))
		}
		for _, pleasureUri := range pleasureUris {
			log.Println("拣到pleasure", string(pleasureUri[1]))
			w.Dispatcher.Set(s.Id, s.Num, string(pleasureUri[1]))
			pleasures = append(pleasures, Pleasure{
				Name:     strconv.Itoa(s.Num),
				Ext:      string(pleasureUri[2]),
				Resource: Resource{Uri: string(pleasureUri[1]), Refer: s.Uri, Id: s.Id, Num: s.Num}})
		}
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

func (w *Worker) PreSave(oriUri string, Id string, textByte []byte) {
	config := w.Dispatcher.Config
	titleReg := regexp.MustCompile(config.TitleReg)
	tagReg := regexp.MustCompile(config.TagReg)
	dateReg := regexp.MustCompile(config.PostTimeReg)
	categoryReg := regexp.MustCompile(config.CategoryReg)
	totalNumReg := regexp.MustCompile(config.TotalPageReg)

	text := string(textByte)
	if len(titleReg.FindStringSubmatch(text)) == 0 {
		log.Fatalln(Id, text)
	}

	title := titleReg.FindStringSubmatch(text)[1]
	date := dateReg.FindStringSubmatch(text)[1]
	category := categoryReg.FindStringSubmatch(text)[1]

	tagsRaw := tagReg.FindAllStringSubmatch(text, -1)
	tags := []string{}
	for _, t := range tagsRaw {
		tags = append(tags, t[1])
	}
	tmp := totalNumReg.FindStringSubmatch(text)
	if len(tmp) == 0 {
		log.Fatalln(Id, oriUri+"没有找到总页数")
	}
	totalNum, err := strconv.Atoi(string(tmp[1]))
	if err != nil {
		log.Fatalln(err)
	}
	//保存图集信息
	w.Dispatcher.Add(Id, oriUri, totalNum, 0, title, category, date, tags)
	//保存网页备份
	err = util.SaveToFile(path.Join("data", w.Dispatcher.Ctx.SiteName, Id, "backup.html"),
		textByte)
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

func UniqueSeeder(ss []Seeder) []Seeder {
	var existUrl map[string]bool
	existUrl = make(map[string]bool)
	var res []Seeder
	for _, s := range ss {
		if _, ok := existUrl[s.Uri]; ok {
			continue
		} else {
			existUrl[s.Uri] = true
			res = append(res, s)
		}
	}
	return res
}
