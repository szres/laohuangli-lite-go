package main

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	kuma "github.com/Nigh/kuma-push"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	scribble "github.com/nanobox-io/golang-scribble"
	tele "gopkg.in/telebot.v3"
)

type testenv struct {
	Token   string `json:"token"`
	AdminID string `json:"adminid"`
	KumaURL string `json:"kumaurl"`
}

var testEnv testenv

type laohuangli struct {
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
}
type laohuangliSlice []laohuangli

var (
	gTimezone    *time.Location = time.FixedZone("CST", 8*60*60)
	gTimeFormat  string         = "2006-01-02 15:04"
	gAdminID     int64
	gKumaPushURL string
	gToken       string

	gStrCompareAlgo *metrics.Jaro
)
var laohuangliList laohuangliSlice
var laohuangliCache map[int64]string

var db *scribble.Driver

func (lhl *laohuangliSlice) init() {
	*lhl = make(laohuangliSlice, 0)
	db.Read("datas", "laohuangli", lhl)
	laohuangliCache = make(map[int64]string)
	db.Read("datas", "cache", &laohuangliCache)
}
func (lhl *laohuangliSlice) add(l laohuangli) {
	*lhl = append(*lhl, l)
	db.Write("datas", "laohuangli", lhl)
}
func (lhl *laohuangliSlice) remove(c string) bool {
	// TODO:
	return false
}
func (lhl *laohuangliSlice) random() string {
	max := big.NewInt(int64(len(*lhl)))
	p, _ := rand.Int(rand.Reader, max)
	n, _ := rand.Int(rand.Reader, max)
	posStr := (*lhl)[p.Int64()].Content
	negStr := (*lhl)[n.Int64()].Content

	if strutil.Similarity(posStr, negStr, gStrCompareAlgo) > 0.95 {
		if p.Cmp(n) > 0 {
			return "今日:\n诸事不宜。请谨慎行事。"
		} else {
			return "今日:\n诸事皆宜。愿好运与你同行。"
		}
	} else {
		return "今日:\n宜" + posStr + "，忌" + negStr + "。"
	}
}

func (lhl *laohuangliSlice) randomResultFromString(s string) string {
	pos := new(big.Int)
	pos.SetBytes(sha1.New().Sum([]byte("positive-" + s)))
	pos.Mod(pos, big.NewInt(int64(len(*lhl))))

	neg := new(big.Int)
	neg.SetBytes(sha1.New().Sum([]byte("negative-" + s)))
	neg.Mod(neg, big.NewInt(int64(len(*lhl))))
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

func (lhl *laohuangliSlice) randomFromDateAndID(t time.Time, id int64) string {
	_, exist := laohuangliCache[id]
	if !exist {
		laohuangliCache[id] = lhl.randomResultFromString(t.In(gTimezone).Format("20060102") + "-" + strconv.FormatInt(id, 10))
		db.Write("datas", "cache", &laohuangliCache)
	}
	return laohuangliCache[id]
}
func (lhl *laohuangliSlice) randomFromRandom(id int64) string {
	_, exist := laohuangliCache[id]
	if !exist {
		laohuangliCache[id] = lhl.random()
		db.Write("datas", "cache", &laohuangliCache)
	}
	return laohuangliCache[id]
}
func (lhl laohuangliSlice) update() {
	ticker := time.NewTicker(1 * time.Second)
	day := time.Now().In(gTimezone).Day()
	for range ticker.C {
		if time.Now().In(gTimezone).Day() != day {
			day = time.Now().In(gTimezone).Day()
			laohuangliCache = make(map[int64]string)
			db.Write("datas", "cache", &laohuangliCache)
		}
	}
}

func kumaPushInit() {
	k := kuma.New(gKumaPushURL)
	k.Start()
}

func init() {
	db, _ = scribble.New("../db", nil)
	laohuangliList.init()
	go laohuangliList.update()

	db.Read("test", "env", &testEnv)
	if testEnv.Token != "" {
		gToken = testEnv.Token
		gAdminID, _ = strconv.ParseInt(testEnv.AdminID, 10, 64)
		gKumaPushURL = testEnv.KumaURL
	} else {
		gToken = os.Getenv("BOT_TOKEN")
		gAdminID, _ = strconv.ParseInt(os.Getenv("BOT_ADMIN_ID"), 10, 64)
		gKumaPushURL = os.Getenv("KUMA_PUSH_URL")
	}
	gStrCompareAlgo = metrics.NewJaro()
	gStrCompareAlgo.CaseSensitive = false
	fmt.Printf("gToken:%s\ngAdminID:%d\ngKumaPushURL:%s\n", gToken, gAdminID, gKumaPushURL)
	kumaPushInit()
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
		Token:  gToken,
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
			Text:  fullName(c.Sender()) + " " + laohuangliList.randomFromRandom(c.Sender().ID),
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
