package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"martin/spider_man/factory"
	"martin/spider_man/global"
)

func initDb() *sql.DB {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatalln("传送带启动失败：", err)
	}
	sql_table := `CREATE TABLE IF NOT EXISTS "done_url" (
"url" VARCHAR(255) NULL,
"created" TIMESTAMP default (datetime('now', 'localtime')));`
	_, err = db.Exec(sql_table)
	if err != nil {
		log.Fatalln("传送带初始化失败：", err)
	}
	return db
}

func main() {
	db := initDb()
	context := global.Context{Db: db}
	//开始
	log.Println("开始")
	manager := factory.Dispatcher{WorkerNum: 1, Ctx: context}
	manager.Start()
}
