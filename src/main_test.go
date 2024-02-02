package main

import (
	"fmt"
	"testing"
	"time"
)

func init() {
	laohuangliList.init()
}
func TestRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(laohuangliList.random())
	}
}
func BenchmarkRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laohuangliList.random()
	}
}
func BenchmarkRandomFromID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laohuangliList.randomFromDateAndID(time.Now(), 12345)
	}
}
