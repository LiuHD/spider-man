package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"martin/spider_man/factory"
	"martin/spider_man/global"
)

func main() {
	db := factory.Init()
	context := global.Context{Db: db}
	//开始
	log.Println("开始")
	manager := factory.Dispatcher{WorkerNum: 1, Ctx: context}
	manager.Start()
}
