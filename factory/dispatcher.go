package factory

import (
	"fmt"
	"log"
	"martin/spider_man/global"
	"strconv"
	"time"
)

type Dispatcher struct {
	Ctx       global.Context
	Config    global.SiteConfig
	WorkerNum int
}

var cs chan Seeder
var ps chan []Pleasure

func (d *Dispatcher) Start() {
	//初始化仓库
	InitKeeper(d.Ctx.SiteName)

	cs = make(chan Seeder, 100)
	ps = make(chan []Pleasure, 100)

	i := 0
	for i < d.WorkerNum {
		w := new(Worker)
		w.PleasureChan = ps
		w.SeederChan = cs
		w.Dispatcher = d
		w.Id = strconv.Itoa(i + 1)
		go w.Run()
		i++
	}
	log.Println(strconv.Itoa(d.WorkerNum) + "个旷工已经开始工作")

	//种子初始化
	seeders := d.GetAllUndone()
	for _, su := range d.Config.ListUrl {
		seeders = append(seeders, Seeder{Resource{Uri: su, Num: 0, Id: ""}})
	}
	for _, s := range seeders {
		go func(seeder Seeder) { cs <- seeder }(s)
	}
	d.panel()

	shutdownSign := time.Tick(1 * time.Second)
	emptyTime := 0
shutdown:
	for {
		select {
		case <-shutdownSign:
			if emptyTime > 30 {
				break shutdown
			} else if emptyTime > 10 {
				//todo 如果有些地址一直返回空，会导致死循环，先加上重试次数的筛选，再开启
				//log.Println("再加一些🐛")
				//seeders := d.GetAllUndone()
				//for _, s := range seeders {
				//	go func(seeder Seeder) {cs <- seeder}(s)
				//}
			}
			if len(cs) == 0 && len(ps) == 0 {
				emptyTime++
			} else {
				emptyTime = 0
			}
		}
	}
	close(cs)
	close(ps)
	log.Println("THE END")
}

func (d *Dispatcher) panel() {
	var internalSign = time.Tick(2 * time.Second)
	go func() {
		for {
			select {
			case <-internalSign:
				fmt.Printf("%d 🐛 %d 🦋 \n\n", len(cs), len(ps))
			}
		}

	}()
}
