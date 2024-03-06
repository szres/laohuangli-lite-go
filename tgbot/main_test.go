package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	scribble "github.com/nanobox-io/golang-scribble"
	"gonum.org/v1/gonum/stat/combin"
)

func init() {
	// db, _ = scribble.New("../db", nil)
	// laoHL.init(db)
	// laoHL.cache.New()
	// laoHL.cache.Save()
}
func TestBase(t *testing.T) {
	fmt.Println(laoHL.cache.Today.String())
}
func TestRandom(t *testing.T) {
	fmt.Println("Length of entries", len(laoHL.entries))
	fmt.Println("Length of entriesBanlanced", len(laoHL.entriesBanlanced))
	for i := 0; i < 10; i++ {
		a, err := laoHL.randomThenDelete()
		b, _ := laoHL.randomThenDelete()
		if err != nil {
			fmt.Printf("Err:[%s]\n", err.Error())
		} else {
			fmt.Printf("宜:[%s]\n忌:[%s]\n", a, b)
		}
	}
	db.Write("datas", "laohuangliBanlanced", laoHL.entriesBanlanced)
}
func TestUserRandom16(t *testing.T) {
	db, _ = scribble.New("../db", nil)
	laoHL.init(db)
	for k, v := range laoHL.cache.Caches {
		fmt.Println(k, v)
	}
	for i := 0; i < 13; i++ {
		userID := strconv.Itoa(1000 + i)
		userName := "TEST USER[" + userID + "]"
		fmt.Println(userName + " " + laoHL.randomToday(int64(1000+i), userName))
	}
}
func TestRegexp(t *testing.T) {
	var list []string = []string{
		`_`, `*`, `[`, `]`, `(`, `)`, `~`, "`", `>`, `#`, `+`, `-`, `=`, `|`, `{`, `}`, `.`, `!`, `\`,
	}
	testStr := "_*[]#12345(test)~测试`>#+-=|{}\n.!\\t"

	prefix := ""
	result := ""
	for _, c := range testStr {
		if strings.ContainsRune(strings.Join(list, ""), rune(c)) && prefix != `\` {
			result += `\` + string(c)
		} else {
			result += string(c)
		}
		prefix = string(c)
	}
	fmt.Println(result)
}
func TestRandomSliceSlice(t *testing.T) {
	getRandomNFromSlice := func(slice []string) (ret []string) {
		// [0, len(slice)^2)
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(slice)*len(slice))))
		// (-len(slice), 0]
		randInt.Sqrt(randInt).Neg(randInt)
		// (0, len(slice)]
		randInt.Add(randInt, big.NewInt(int64(len(slice))))
		// pick (0, len(slice)] elements from slice
		list := combin.Combinations(len(slice), int(randInt.Int64()))
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(list))))
		listPick := list[randInt.Int64()]
		for _, k := range listPick {
			ret = append(ret, slice[k])
		}
		return
	}
	for i := 0; i < 10; i++ {
		fmt.Println(getRandomNFromSlice([]string{"a", "b", "c", "d", "e", "f"}))
	}
}
func BenchmarkRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		laoHL.randomThenDelete()
	}
}

// func BenchmarkRandomStable(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		id := int64(i)
// 		laoHL.randomStable(time.Now().In(gTimezone).Format("20060102") + "-" + strconv.FormatInt(id, 10))
// 	}
// }
