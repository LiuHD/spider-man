package dispatcher

import (
	"fmt"
	"log"
	"martin/spider_man/worker"
	"strconv"
	"time"
)

type Dispatcher struct {
	WorkerNum int
}

var cs chan worker.Seeder
var ps chan []worker.Pleasure
var doneUri map[string]bool

func(d *Dispatcher) Start() {
	cs = make(chan worker.Seeder, 10)
	ps = make(chan []worker.Pleasure, 10)
	doneUri = make(map[string]bool)

	i := 0
	for i < d.WorkerNum {
		w := new(worker.Worker)
		w.PleasureChan = ps
		w.SeederChan = cs
		w.DoneUri = doneUri
		w.Id = strconv.Itoa(i + 1)
		w.Run()
		i ++
	}
	log.Println(strconv.Itoa(d.WorkerNum) + "个旷工已将开始工作")
	cs <- worker.Seeder{worker.Resource{Uri:"https://www.mzitu.com"}}
	d.panel()

	shutdownSign := time.Tick(1 * time.Second)
	emptyTime := 0
	shutdown:
	for {
		select {
		case <-shutdownSign:
			if emptyTime > 10 {
				break shutdown
			}
			if len(cs) == 0 && len(ps) == 0 {
				emptyTime ++
			} else {
				emptyTime = 0
			}
		}
	}
}

func(d *Dispatcher) panel() {
	var internalSign = time.Tick(2 * time.Second)
	go func() {
		select {
			case <-internalSign:
				fmt.Printf("seeder num: %d\npleasure num: %d\ndone uri num: %d\n\n", len(cs), len(ps), len(doneUri))

		}
	}()
}
