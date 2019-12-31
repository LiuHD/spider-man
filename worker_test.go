package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"martin/spider_man/factory"
	"martin/spider_man/util"
	"os"
	"path"
	"regexp"
	"testing"
)

func TestPick(t *testing.T) {
	w := new(factory.Worker)
	var text []byte
	file, err := os.Open("data/tmp/example3.html")
	if err != nil {
		t.Fatal(err)
	}
	text, err = ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(w.Pick("www.baidu.com", "", "", text))
	t.Log("\n")
}

func TestSaveToFile(t *testing.T) {
	t.Log(util.SaveToFile("data/2/1/1/1/11.txt", []byte("hello world!")))
}

func TestDigAndSave(t *testing.T) {
	w := factory.Worker{Id: "1"}
	p := factory.Pleasure{
		Resource: factory.Resource{
			Refer: "https://www.mzitu.com/201884/4",
			Uri:   "https://i5.meizitu.net/2016/02/29z13.jpg"}}
	entity, err := w.Dig(p.Resource)
	if err != nil {
		t.Log("旷工"+w.Id+"来报，搬运出错了，", err)
	}
	err = util.SaveToFile("data/2.jpg", entity)
	if err != nil {
		log.Println("旷工"+w.Id+"来报，储存出错了，", err)
	}
	t.Log("已保存！")
}

func TestPreSave(t *testing.T) {
	w := factory.Worker{Id: "1"}
	var text []byte
	file, err := os.Open("data/tmp/example3.html")
	if err != nil {
		t.Fatal(err)
	}
	text, err = ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	w.PreSave("1234", text)
}

func TestFun(t *testing.T) {
	pleasureReg := regexp.MustCompile(`https://i\d{0,2}\.meizitu\.net/(\d.*?)\.(jpg|jpeg|png|bmp|gif)`)
	if firstPage := pleasureReg.FindStringSubmatch("https://i5.meizitu.net/2019/09/07d05.jpg"); firstPage != nil {
		fmt.Println(firstPage)
		fmt.Println(path.Join("data", firstPage[1]) + "." + firstPage[2])
	}
}
