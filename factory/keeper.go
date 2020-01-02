package factory

import (
	"database/sql"
	"log"
	"path"
	"strconv"
	"time"
)

func Init() *sql.DB {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatalln("传送带启动失败：", err)
	}
	sql_table := `CREATE TABLE IF NOT EXISTS "gallery" (
"gallery_id" varchar NOT NULL,
"url" varchar NOT NULL,
"total_num" integer NOT NULL DEFAULT 0,
"done_num" integer NOT NULL DEFAULT 0,
"created" TIMESTAMP,
"updated" TIMESTAMP default (datetime('now', 'localtime')))`
	table2_sql := `CREATE TABLE IF NOT EXISTS "pleasure" (
"gallery_id" varchar not null,
"num" integer not null default 0,
"done" integer not null default 0,
"created" timestamp ,
"updated" timestamp default (datetime('now', 'localtime')))`
	_, err = db.Exec(sql_table)
	if err != nil {
		log.Fatalln("传送带初始化失败：", err)
	}
	_, err = db.Exec(table2_sql)
	if err != nil {
		log.Fatalln("传送带初始化失败：", err)
	}
	return db
}

func (d *Dispatcher) Exist(Id string) bool {
	stmt, err := d.Ctx.Db.Prepare("SELECT * FROM gallery where gallery_id = ? limit 1")
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

func (d *Dispatcher) Add(Id string, uri string, totalNum int, doneNum int) {
	stmt, err := d.Ctx.Db.Prepare("INSERT INTO gallery(gallery_id, url, total_num, done_num, created) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(Id, uri, totalNum, doneNum, time.Now().Unix())
	if err != nil {
		log.Fatalln("写入报错", err)
	}
	stmt1, err := d.Ctx.Db.Prepare("INSERT INTO pleasure(gallery_id, num, created) values(?, ?, ?)")
	defer stmt1.Close()
	num := 0
	for num < totalNum {
		num++
		stmt1.Exec(Id, num, time.Now().Unix())
	}
}

func (d *Dispatcher) Ensure() {
	res, err := d.Ctx.Db.Query("SELECT gallery_id, total_num FROM gallery limit")
	if err != nil {
		log.Fatalln(err)
	}
	var totalNum int
	var GalleryId string
	for res.Next() {
		res.Scan(&GalleryId, &totalNum)
		//log.Fatalln(GalleryId, totalNum)
		i := 1
		for i <= totalNum {
			stmt, err := d.Ctx.Db.Prepare("SELECT * FROM pleasure where gallery_id = ? and num = ? limit 1")
			if err != nil {
				log.Fatalln(err)
			}
			defer stmt.Close()

			rows, err := stmt.Query(GalleryId, i)
			defer rows.Close()

			if err != nil {
				log.Fatalln(err)
			}
			//log.Println(GalleryId, i, rows.Next())
			if !rows.Next() {
				log.Println("补", GalleryId, i)
				stmt1, _ := d.Ctx.Db.Prepare("INSERT INTO pleasure(gallery_id, num, created) values(?, ?, ?)")
				defer stmt1.Close()
				stmt1.Exec(GalleryId, i, time.Now().Unix())
			}
			i++
		}
	}
}

func (d *Dispatcher) SetDone(num int, Id string) {
	stmt, err := d.Ctx.Db.Prepare("UPDATE pleasure set done = 1 where gallery_id = ? and num = ?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(Id, num)
	if err != nil {
		log.Fatalln("写入完成态报错1", err)
	}

	stmt1, err := d.Ctx.Db.Prepare("UPDATE gallery set done_num = (done_num + 1) where gallery_id = ?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt1.Close()
	_, err = stmt1.Exec(Id)
	if err != nil {
		log.Fatalln("写入完成态报错2", err)
	}
}

func (d *Dispatcher) GetAllUndone() []Seeder {
	stmt, err := d.Ctx.Db.Prepare("SELECT gallery_id, num FROM pleasure where done = 0")
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
		var GalleryId string
		var Num int
		err = rows.Scan(&GalleryId, &Num)
		if err != nil {
			log.Fatalln(err)
		}
		seeders = append(seeders, Seeder{Resource{Id: GalleryId, Num: Num, Uri: d.genSeederUri(GalleryId, Num)}})
	}

	return seeders
}

func (d *Dispatcher) genSeederUri(Id string, Num int) string {
	if Num == 1 {
		return "https://" + path.Join("www.mzitu.com", Id)
	}
	return "https://" + path.Join("www.mzitu.com", Id, strconv.Itoa(Num))
}
