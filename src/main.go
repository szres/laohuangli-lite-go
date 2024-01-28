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

var gTimezone *time.Location = time.FixedZone("CST", 8*60*60)
var laohuangliList []laohuangli
var db *scribble.Driver

func init() {
	db, _ = scribble.New("../db", nil)
	db.Read("datas", "laohuangli", &laohuangliList)
}

func saveLaohuangli() {
	db.Write("datas", "laohuangli", &laohuangliList)
}

var b *tele.Bot

func fullName(u *tele.User) string {
	if u.FirstName == "" || u.LastName == "" {
		return u.FirstName + u.LastName
	}
	return u.FirstName + " " + u.LastName
}

func main() {
	fmt.Println("老黄历启动！")
	pref := tele.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 5 * time.Second},
	}
	var err error
	b, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})
	b.Handle("/help", func(c tele.Context) error {
		return cmdOnChat(c)
	})
	b.Handle("/start", func(c tele.Context) error {
		return cmdOnChat(c)
	})
	b.Handle("/nominate", func(c tele.Context) error {
		return cmdOnChat(c)
	})
	b.Handle("/list", func(c tele.Context) error {
		return cmdOnChat(c)
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		return msgOnChat(c)
	})

	b.Handle(tele.OnQuery, func(c tele.Context) error {

		pos := new(big.Int)
		pos.SetBytes(sha1.New().Sum([]byte("positive-" + time.Now().In(gTimezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		pos.Mod(pos, big.NewInt(int64(len(laohuangliList))))

		neg := new(big.Int)
		neg.SetBytes(sha1.New().Sum([]byte("negative-" + time.Now().In(gTimezone).Format("20060102") + "-" + strconv.FormatInt(c.Sender().ID, 10))))
		neg.Mod(neg, big.NewInt(int64(len(laohuangliList))))

		results := make(tele.Results, 0)
		if pos.Int64() != neg.Int64() {
			results = append(results, &tele.ArticleResult{
				Title: "今日我的老黄历",
				Text:  fullName(c.Sender()) + " 今日:\n宜" + laohuangliList[pos.Int64()].Content + "，忌" + laohuangliList[neg.Int64()].Content + "。",
			})
		} else {
			if pos.Int64()%2 == 0 {
				results = append(results, &tele.ArticleResult{
					Title: "今日我的老黄历",
					Text:  fullName(c.Sender()) + " 今日:\n诸事不宜。请谨慎行事。",
				})
			} else {
				results = append(results, &tele.ArticleResult{
					Title: "今日我的老黄历",
					Text:  fullName(c.Sender()) + " 今日:\n诸事皆宜。愿好运与你同行。",
				})
			}
		}
		for _, v := range nominations {
			if v.NominatorID == c.Sender().ID {
				results = append(results, buildVotes(v))
			}
		}

		return c.Answer(&tele.QueryResponse{
			Results:           results,
			CacheTime:         3,
			IsPersonal:        true,
			SwitchPMText:      "提名新词条",
			SwitchPMParameter: "nominate",
		})
	})

	fmt.Println("上线！")
	b.Start()
}
