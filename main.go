package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"martin/spider_man/factory"
	"martin/spider_man/global"
	"os"
	"strconv"
)

func main() {
	db := factory.Init()
	context := global.Context{Db: db}
	//开始
	log.Println("开始")
	s := os.Args
	var workerNum = 1
	var err error
	if len(s) > 1 {
		workerNum, err = strconv.Atoi(s[1])
		if err != nil {
			log.Fatalln(err)
		}
	}
	manager := factory.Dispatcher{WorkerNum: workerNum, Ctx: context}
	manager.Start()
}
