package main

import (
	"fmt"
	"testing"

	scribble "github.com/nanobox-io/golang-scribble"
)

func init() {
	db, _ = scribble.New("../db", nil)
	laoHL.init(db)
}
func TestRandom(t *testing.T) {
	fmt.Println("Length of entries", len(laoHL.entries))
	fmt.Println("Length of entriesBanlanced", len(laoHL.entriesBanlanced))
	for i := 0; i < 10; i++ {
		a, b, err := laoHL.random()
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		fmt.Printf("宜:[%s] 忌:[%s] Err:[%s]\n", a, b, errStr)
	}
	db.Write("datas", "laohuangliBanlanced", laoHL.entriesBanlanced)
}
func BenchmarkRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laoHL.random()
	}
}

// func BenchmarkRandomStable(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		id := int64(i)
// 		laoHL.randomStable(time.Now().In(gTimezone).Format("20060102") + "-" + strconv.FormatInt(id, 10))
// 	}
// }
