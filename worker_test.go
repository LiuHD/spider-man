package main

import (
	"log"
	"testing"
)

func TestFun(t *testing.T) {
	var existUrl map[string]bool
	existUrl = make(map[string]bool)
	var res []string
	for _, s := range []string{"1", "2", "1", "2", "3", "4"} {
		if _, ok := existUrl[s]; ok {
			continue
		} else {
			existUrl[s] = true
			res = append(res, s)
		}
	}
	log.Fatalln(res)
}
