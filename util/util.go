package util

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func DecodeGzip(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return []byte{}, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func DecodeDeflate(in []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(in))
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func CheckErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func SaveToFile(pathStr string, content []byte) (err error) {
	dir := path.Dir(pathStr)
	name := path.Base(pathStr)
	pathStr = path.Join(dir, name)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return fmt.Errorf("创建文件夹出错:%s", err.Error())
	}
	var file *os.File
	file, err = os.OpenFile(pathStr, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(content)
	return nil
}