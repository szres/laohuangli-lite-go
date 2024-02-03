package main

import (
	"fmt"
	"testing"
	"time"
)

func init() {
	laohuangliList.init()
	laohuangliListBanlanced = laohuangliList.banlance()
}
func TestRandom(t *testing.T) {
	fmt.Println("Length of laohuangliList", len(laohuangliList))
	fmt.Println("Length of laohuangliListBanlanced", len(laohuangliListBanlanced))
	for i := 0; i < 10; i++ {
		fmt.Println(laohuangliListBanlanced.random())
	}
	db.Write("datas", "laohuangliBanlanced", laohuangliListBanlanced)
}
func BenchmarkRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laohuangliListBanlanced.random()
	}
}
func BenchmarkRandomFromID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laohuangliListBanlanced.randomFromDateAndID(time.Now(), 12345)
	}
}
