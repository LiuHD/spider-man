package main

import (
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"martin/spider_man/factory"
	"martin/spider_man/global"
	"regexp"
)

func main() {
	var siteName string
	var workerNum int
	flag.IntVar(&workerNum, "w", 1, "线程数")
	flag.StringVar(&siteName, "n", "mzitu", "网站名")
	flag.Parse()

	context := global.Context{
		WorkerNum: workerNum,
		SiteName:  siteName,
	}
	configs := global.Configs{}
	configs["mei101"] = global.SiteConfig{
		TotalPageReg: `yaomei.max_page = (\d+?);`,
		NextPageReg:  "",
		SeederReg:    `https://www.mei101.com/\w+?/(\d+?).html`,
		TitleReg:     `<h1>(.*?)<span class="post_title_topimg`,
		CategoryReg:  `rel="category tag">(.*?)</a>`,
		TagReg:       `/tag/[^>]*?>(.+?)</a>`,
		PostTimeReg:  `<span>发布于</span>(.+?)\s+</div>`,
		PleasureReg:  `<img src="(.+?\.(jpg))"\s*?alt`,
		ListUrl: []string{
			"https://www.mei101.com/views/page/4",
			"https://www.mei101.com/views/page/5",
			"https://www.mei101.com/views/page/6",
		},
		PageUrlGenerator: func(originUrl string, id string, num int) string {
			reg := regexp.MustCompile(`https://www.mei101.com/(.+?)/(\d+?).html`)
			res := reg.FindStringSubmatch(originUrl)
			if res == nil {
				log.Fatalln("没有检查到合适的源网址", originUrl)
			}
			return fmt.Sprintf("https://www.mei101.com/%s/%s/%d.html", res[1], id, num)
		},
		IdGetter: func(url string) string {
			reg := regexp.MustCompile(`https://www.mei101.com/(.+?)/(\d+?).html`)
			res := reg.FindStringSubmatch(url)
			if res == nil {
				log.Fatalln("没有检查到合适的源网址", url)
			}
			return res[2]
		},
	}

	if _, ok := configs[siteName]; !ok {
		log.Fatalf("没有找到%s的配置\n", siteName)
	}

	manager := factory.Dispatcher{WorkerNum: workerNum, Ctx: context, Config: configs[siteName]}
	manager.Start()
}
