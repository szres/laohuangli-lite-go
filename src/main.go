package main

import (
	"crypto/sha1"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
	tele "gopkg.in/telebot.v3"
)

type laohuangli struct {
	Content string `json:"content"`
}

var laohuangliList []laohuangli
var db *scribble.Driver

func init() {
	db, _ = scribble.New("./db", nil)
	db.Read("datas", "laohuangli", &laohuangliList)
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})

	b.Handle(tele.OnQuery, func(c tele.Context) error {
		timezone, _ := time.LoadLocation("Asia/Shanghai")
		pos := new(big.Int)
		pos.SetBytes(sha1.New().Sum([]byte("positive-" + time.Now().In(timezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		pos.Mod(pos, big.NewInt(int64(len(laohuangliList))))

		neg := new(big.Int)
		neg.SetBytes(sha1.New().Sum([]byte("negative-" + time.Now().In(timezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		neg.Mod(neg, big.NewInt(int64(len(laohuangliList))))

		return c.Answer(&tele.QueryResponse{
			Results: tele.Results{&tele.ArticleResult{
				Title: "今日我的老黄历",
				Text:  c.Sender().Username + " 今日:\n宜" + laohuangliList[pos.Int64()].Content + "\n忌" + laohuangliList[neg.Int64()].Content,
			}},
			CacheTime: 15,
		})
	})

	b.Start()
}
