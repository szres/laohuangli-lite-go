package main

import (
	"fmt"
	"testing"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

func similarity() []float64 {
	// method := metrics.NewSorensenDice()
	// method := metrics.NewHamming()
	// method := metrics.NewJaccard()
	method := metrics.NewJaro()
	// method := metrics.NewJaroWinkler()

	method.CaseSensitive = false
	similarity := make([]float64, 0)
	similarity = append(similarity, strutil.Similarity("新老黄历测试条目 - 001", "新老黄历测试条", method))
	similarity = append(similarity, strutil.Similarity("新老黄历测试条目 - 001", "新老黄历测试", method))
	similarity = append(similarity, strutil.Similarity("新老黄历测试条目 - 001", "新老黄历测", method))
	similarity = append(similarity, strutil.Similarity("新老黄历测试条目 - 001", "旧老黄历目", method))
	return similarity
}
func TestSimilarity(t *testing.T) {
	fmt.Println(similarity())
}
func BenchmarkSimilarity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		similarity()
	}
}
