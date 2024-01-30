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
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
}
type laohuangliSlice []laohuangli

var (
	gTimezone   *time.Location = time.FixedZone("CST", 8*60*60)
	gTimeFormat string         = "2006-01-02 15:04"
	gAdminID    int64
)
var laohuangliList laohuangliSlice
var laohuangliValidLength int64
var db *scribble.Driver

func (lhl *laohuangliSlice) init() {
	*lhl = make(laohuangliSlice, 0)
	db.Read("datas", "laohuangli", lhl)
	laohuangliValidLength = int64(len(*lhl))
}
func (lhl *laohuangliSlice) add(l laohuangli) {
	*lhl = append(*lhl, l)
	db.Write("datas", "laohuangli", lhl)
}
func (lhl *laohuangliSlice) remove(c string) bool {
	// TODO:
	return false
}
func (lhl *laohuangliSlice) randomResultFromString(s string) string {
	pos := new(big.Int)
	pos.SetBytes(sha1.New().Sum([]byte("positive-" + s)))
	pos.Mod(pos, big.NewInt(laohuangliValidLength))

	neg := new(big.Int)
	neg.SetBytes(sha1.New().Sum([]byte("negative-" + s)))
	neg.Mod(neg, big.NewInt(laohuangliValidLength))
	if pos.Int64() != neg.Int64() {
		return "今日:\n宜" + (*lhl)[pos.Int64()].Content + "，忌" + (*lhl)[neg.Int64()].Content + "。"
	} else {
		if pos.Int64()%2 == 0 {
			return "今日:\n诸事不宜。请谨慎行事。"
		} else {
			return "今日:\n诸事皆宜。愿好运与你同行。"
		}
	}
}
func (lhl laohuangliSlice) update() {
	minute := time.NewTicker(1 * time.Minute)
	day := time.Now().In(gTimezone).Day()
	for range minute.C {
		if time.Now().In(gTimezone).Day() != day {
			day = time.Now().In(gTimezone).Day()
			laohuangliValidLength = int64(len(laohuangliList))
		}
	}
}

func init() {
	db, _ = scribble.New("../db", nil)
	laohuangliList.init()
	go laohuangliList.update()
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
	gAdminID, _ = strconv.ParseInt(os.Getenv("BOT_ADMIN_ID"), 10, 64)
	pref := tele.Settings{
		// Token:  os.Getenv("BOT_TOKEN"),
		Token:  "252481040:AAHmOfe5_eztE0DckaCjpwUXILvWUuASHBY",
		Poller: &tele.LongPoller{Timeout: 5 * time.Second},
	}
	var err error
	b, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, s := range []string{
		"/help", "/start", "/nominate", "/list", "/listall", "/forcereadlocal",
	} {
		b.Handle(s, func(c tele.Context) error {
			return cmdInChatHandler(c)
		})
	}

	b.Handle(tele.OnText, func(c tele.Context) error {
		return msgInChatHandler(c)
	})

	mk := &tele.ReplyMarkup{ResizeKeyboard: true}
	voteApproveBtn := mk.Data("赞成", "voteApproveBtn")
	voteRefuseBtn := mk.Data("反对", "voteRefuseBtn")
	deleteBtn := mk.Data("删除", "deleteBtn")
	b.Handle(&voteApproveBtn, voteApprove())
	b.Handle(&voteRefuseBtn, voteRefuse())
	b.Handle(&deleteBtn, msgDelete())

	b.Handle(tele.OnQuery, func(c tele.Context) error {
		results := make(tele.Results, 0)
		for _, v := range nominations {
			if v.NominatorID == c.Sender().ID {
				results = append(results, buildVotes(v))
			}
		}
		results = append(results, &tele.ArticleResult{
			Title: "今日我的老黄历",
			Text:  fullName(c.Sender()) + " " + laohuangliList.randomResultFromString(time.Now().In(gTimezone).Format("20060102")+"-"+strconv.FormatInt(c.Sender().ID, 10)),
		})
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
