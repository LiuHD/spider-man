package factory

import (
	"fmt"
	"log"
	"martin/spider_man/global"
	"strconv"
	"time"
)

const LIST_URI = "https://www.mzitu.com/all"
const HOME_URI = "https://www.mzitu.com"

type Dispatcher struct {
	Ctx       global.Context
	WorkerNum int
}

var cs chan Seeder
var ps chan []Pleasure

func (d *Dispatcher) Start() {
	//d.Ensure()
	//log.Fatalln("结束了")
	cs = make(chan Seeder, 10)
	ps = make(chan []Pleasure, 10)

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
	seeders := d.GetAllUndone()
	for _, s := range seeders {
		cs <- s
	}
	cs <- Seeder{Resource{Uri: LIST_URI}}
	d.panel()

	shutdownSign := time.Tick(1 * time.Second)
	emptyTime := 0
shutdown:
	for {
		select {
		case <-shutdownSign:
			if emptyTime > 30 {
				break shutdown
			}
			if len(cs) == 0 && len(ps) == 0 {
				emptyTime++
			} else {
				emptyTime = 0
			}
		}
	}
	log.Println("THE END")
}

func (d *Dispatcher) panel() {
	var internalSign = time.Tick(2 * time.Second)
	go func() {
		select {
		case <-internalSign:
			//todo
			fmt.Printf("seeder num: %d\npleasure num: %d\ndone uri num: %d\n\n", len(cs), len(ps), 0)

		}
	}()
}
