package main

import (
	"fmt"
	"log"
	"martin/spider_man/dispatcher"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"martin/spider_man/util"
)

func initDb() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatalln("传送带启动失败：", err)
	}
	sql_table := `CREATE TABLE IF NOT EXISTS "userinfo" ("uid" INTEGER ,"username" VARCHAR(64) NULL, "departname" VARCHAR(64) NULL, "created" TIMESTAMP default (datetime('now', 'localtime')));`
	_, err = db.Exec(sql_table)
	if err != nil {
		log.Fatalln("传送带初始化失败：", err)
	}
	stmt, err := db.Prepare("INSERT INTO userinfo(username, departname)  values(?, ?)")
	util.CheckErr(err)
	res, err := stmt.Exec("刘汇东", 1)
	util.CheckErr(err)
	fmt.Println(res)
}

func main() {
	initDb()
}

func main1() {
	//开始
	log.Println("开始")
	manager := dispatcher.Dispatcher{WorkerNum:1}
	manager.Start()
}


