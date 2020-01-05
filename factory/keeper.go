package factory

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

var db *sql.DB

var galleryTableName, pleasureTableName, tagTableName string

func InitKeeper(siteName string) {
	galleryTableName = siteName + "_gallery"
	pleasureTableName = siteName + "_pleasure"
	tagTableName = siteName + "_tag"

	var err error
	db, err = sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatalln("传送带启动失败：", err)
	}

	gallery_sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
"id" varchar NOT NULL,
"url" varchar NOT NULL,
"total_num" integer NOT NULL DEFAULT 0,
"done_num" integer NOT NULL DEFAULT 0,
"title" varchar Null,
"category" varcha NULL,
"posted" varchar null,
"created" TIMESTAMP,
"updated" TIMESTAMP default (datetime('now', 'localtime')))`, galleryTableName)
	pleasure_sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
"url" varchar not null,
"referrer_url" varchar not null,
"gallery_id" varchar not null,
"num" integer not null default 0,
"done" integer not null default 0,
"created" timestamp,
"updated" timestamp default (datetime('now', 'localtime')))`, pleasureTableName)
	tag_sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
"gallery_id" varchar not null,
"tag" varchar NOT NULL,
"created" timestamp ,
"updated" timestamp default (datetime('now', 'localtime')))`, tagTableName)
	_, err = db.Exec(gallery_sql)
	if err != nil {
		log.Fatalln("宝罐初始化失败：", err)
	}
	_, err = db.Exec(pleasure_sql)
	if err != nil {
		log.Fatalln("宝贝初始化失败：", err)
	}
	_, err = db.Exec(tag_sql)
	if err != nil {
		log.Fatalln("标签初始化失败：", err)
	}
}

func (d *Dispatcher) Exist(Id string) bool {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT * FROM %s where id = ? limit 1", galleryTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(Id)
	defer rows.Close()

	if err != nil {
		log.Fatalln(err)
	}
	return rows.Next()
}

func (d *Dispatcher) PleasureExist(Id string, num int) bool {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT * FROM %s where gallery_id = ? and num = ? limit 1", pleasureTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(Id, num)
	defer rows.Close()

	if err != nil {
		log.Fatalln(err)
	}
	return rows.Next()
}

func (d *Dispatcher) Add(Id string, uri string, totalNum int, doneNum int, title, category, posted string, tags []string) {
	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(id, url, total_num, done_num, title, category, posted, created) VALUES(?, ?, ?, ?, ?, ?, ?, ?)", galleryTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(Id, uri, totalNum, doneNum, title, category, posted, time.Now().Unix())
	if err != nil {
		log.Fatalln("写入gallery", err)
	}
	stmt1, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(gallery_id, num, created, url, referrer_url) values(?, ?, ?, ?, ?)", pleasureTableName))
	defer stmt1.Close()
	pageUrlGenerator := d.Config.PageUrlGenerator
	if pageUrlGenerator != nil {
		num := 0
		for num < totalNum {
			num++
			stmt1.Exec(Id, num, time.Now().Unix(), "", pageUrlGenerator(uri, Id, num))
		}
	}
	if err != nil {
		log.Fatalln("写入pleasure", err)
	}
	stmt2, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(gallery_id, tag, created) values(?, ?, ?)", tagTableName))
	defer stmt2.Close()
	for _, tag := range tags {
		stmt2.Exec(Id, tag, time.Now().Unix())
	}
	if err != nil {
		log.Fatalln("写入tag", err)
	}
}

func (d *Dispatcher) PleasureAdd(referUrl string, Id string, uri string, num int) {
	stmt1, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(gallery_id, num, created, url, referrer_url) values(?, ?, ?, ?, ?)", pleasureTableName))
	defer stmt1.Close()
	stmt1.Exec(Id, num, time.Now().Unix(), uri, referUrl)
	if err != nil {
		log.Fatalln("写入pleasure", err)
	}
}

func (d *Dispatcher) SetDone(num int, Id string) {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s set done = 1 where gallery_id = ? and num = ?", pleasureTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(Id, num)
	if err != nil {
		log.Fatalln("写入完成态1", err)
	}

	stmt1, err := db.Prepare(fmt.Sprintf("UPDATE %s set done_num = (select count(*) from %s where gallery_id = ? and done = 1) where id = ?", galleryTableName, pleasureTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt1.Close()
	_, err = stmt1.Exec(Id, Id)
	if err != nil {
		log.Fatalln("写入完成态2", err)
	}
}

func (d *Dispatcher) Set(Id string, num int, url string) {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s set url = ? where gallery_id = ? and num = ?", pleasureTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(url, Id, num)
	if err != nil {
		log.Fatalln("完善pleasure地址失败", err)
	}
}

func (d *Dispatcher) GetAllUndone() []Seeder {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT url, referrer_url, gallery_id, num FROM %s where done = 0", pleasureTableName))
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
	}
	var seeders []Seeder
	for rows.Next() {
		var GalleryId, Url, ReferrerUrl string
		var Num int
		err = rows.Scan(&Url, &ReferrerUrl, &GalleryId, &Num)
		if err != nil {
			log.Fatalln(err)
		}
		seeders = append(seeders, Seeder{Resource{Id: GalleryId, Num: Num, Uri: ReferrerUrl}})
	}

	return seeders
}
