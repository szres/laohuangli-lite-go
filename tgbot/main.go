package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	kuma "github.com/Nigh/kuma-push"
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

type entry struct {
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
}
type laohuangliTemplate struct {
	Desc   string   `json:"desc"`
	Values []string `json:"values"`
}
type laohuangliResult struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}

var (
	gTimezone    *time.Location = time.FixedZone("CST", 8*60*60)
	gTimeFormat  string         = "2006-01-02 15:04"
	gAdminID     int64
	gKumaPushURL string
	gToken       string

	gStrCompareAlgo *metrics.Jaro
)

var db *scribble.Driver

func init() {
	db, _ = scribble.New("../db", nil)

	laoHL.init(db)
	laoHL.start()

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
	k := kuma.New(gKumaPushURL)
	k.Start()
	go func() {
		http.Handle("/", http.FileServer(http.Dir("../db/datas")))
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			panic(err)
		}
	}()
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

	for _, s := range chatCMD {
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
			Title: "今日众生老黄历",
			Text:  "今天是" + time.Now().In(gTimezone).Format("2006年01月02日") + "。\n" + laoHL.cache.Today.String(),
		})
		results = append(results, &tele.ArticleResult{
			Title: "今日我的老黄历",
			Text:  fullName(c.Sender()) + " " + laoHL.randomToday(c.Sender().ID, fullName(c.Sender())),
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
	go b.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	b.Stop()
	fmt.Println("由于即将关闭，进行数据备份")
	laoHL.save()
	<-time.After(time.Second * 1)
	fmt.Println("下线！")
}
