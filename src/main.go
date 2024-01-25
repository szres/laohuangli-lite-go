package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
	tele "gopkg.in/telebot.v3"
)

type laohuangli struct {
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
}

var laohuangliList []laohuangli
var db *scribble.Driver

func init() {
	db, _ = scribble.New("./db", nil)
	db.Read("datas", "laohuangli", &laohuangliList)
}

func saveLaohuangli() {
	db.Write("datas", "laohuangli", &laohuangliList)
}

func main() {
	fmt.Println("老黄历启动！")
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
	b.Handle("/start", func(c tele.Context) error {
		MsgOnChat(c)
		return c.Send("如果要提名新词条请发送 /nominate")
	})
	b.Handle("/nominate", func(c tele.Context) error {
		MsgOnChat(c)
		return c.Send("请输入你要提名的词条内容：")
	})
	b.Handle("/fnominate", func(c tele.Context) error {
		MsgOnChat(c)
		return c.Send("使用强制提名模式，将会跳过查重直接进入提名，请确认你的提名确实有效，然后输入你要提名的词条内容：")
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		MsgOnChat(c)
		return nil
	})
	b.Handle(tele.OnPoll, func(c tele.Context) error {
		fmt.Printf("c.Poll(): %v\n", c.Poll())
		return nil
	})
	b.Handle(tele.OnPollAnswer, func(c tele.Context) error {
		fmt.Printf("c.PollAnswer(): %v\n", c.PollAnswer())
		return nil
	})
	b.Handle(tele.OnQuery, func(c tele.Context) error {
		timezone := time.FixedZone("CST", 8*60*60)

		pos := new(big.Int)
		pos.SetBytes(sha1.New().Sum([]byte("positive-" + time.Now().In(timezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		pos.Mod(pos, big.NewInt(int64(len(laohuangliList))))

		neg := new(big.Int)
		neg.SetBytes(sha1.New().Sum([]byte("negative-" + time.Now().In(timezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		neg.Mod(neg, big.NewInt(int64(len(laohuangliList))))

		return c.Answer(&tele.QueryResponse{
			Results: tele.Results{
				&tele.ArticleResult{
					Title: "今日我的老黄历",
					Text:  c.Sender().FirstName + c.Sender().LastName + " 今日:\n宜" + laohuangliList[pos.Int64()].Content + "，忌" + laohuangliList[neg.Int64()].Content + "。",
				}},
			CacheTime:         15,
			IsPersonal:        true,
			SwitchPMText:      "提名新词条",
			SwitchPMParameter: "nominate",
		})
	})

	fmt.Println("上线！")
	b.Start()
}
