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
	WorkerNum int
}

var cs chan Seeder
var ps chan []Pleasure

func (d *Dispatcher) SetDone(uri string) {
	stmt, err := d.Ctx.Db.Prepare("INSERT INTO done_url(url) values (?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(uri)
	if err != nil {
		log.Fatalln("写入报错", err)
	}
}

func (d *Dispatcher) GetDone(uri string) bool {
	if uri == "https://www.mzitu.com" || uri == "https://www.mzitu.com/all" {
		return false
	}
	stmt, err := d.Ctx.Db.Prepare("SELECT * FROM done_url where url = ? limit 1")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(uri)
	defer rows.Close()

	if err != nil {
		log.Fatalln(err)
	}
	return rows.Next()
}

func (d *Dispatcher) Start() {
	cs = make(chan Seeder, 10)
	ps = make(chan []Pleasure, 10)

	i := 0
	for i < d.WorkerNum {
		w := new(Worker)
		w.PleasureChan = ps
		w.SeederChan = cs
		w.Dispatcher = d
		w.Id = strconv.Itoa(i + 1)
		w.Run()
		i++
	}
	log.Println(strconv.Itoa(d.WorkerNum) + "个旷工已经开始工作")
	cs <- Seeder{Resource{Uri: "https://www.mzitu.com/all"}}
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
