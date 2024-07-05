package main

import (
	"crypto/rand"
	"math/big"
	"time"
)

func PP(n int64) bool {
	if n > 100 {
		return true
	}
	if n <= 0 {
		return false
	}
	randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(1000)))
	return randInt.Cmp(big.NewInt(n*10)) <= 0
}

func ingressStr() string {
	now := time.Now()
	if now.Weekday() == time.Saturday && now.Day() <= 7 {
		// IFS day
		if PP(17) {
			return "参加IFS"
		}
	}
	if now.Weekday() == time.Tuesday {
		// Double AP day
		if PP(13) {
			return "刷AP"
		}
	}
	if now.Weekday() == time.Sunday && now.Day() > 7 && now.Day() <= 14 {
		// ISS day
		if PP(11) {
			return "出门做一排任务"
		}
	}
	return ""
}
