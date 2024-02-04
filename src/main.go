package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	kuma "github.com/Nigh/kuma-push"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/valyala/fasttemplate"
	tele "gopkg.in/telebot.v3"
)

type testenv struct {
	Token   string `json:"token"`
	AdminID string `json:"adminid"`
	KumaURL string `json:"kumaurl"`
}

var testEnv testenv

type laohuangli struct {
	// 原始词条
	entries []entry
	// 频次均衡后的词条
	entriesBanlanced []entry
	templates        map[string]laohuangliTemplate
	cache            laohuangliCache
}
type entry struct {
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
}
type laohuangliTemplate struct {
	Desc   string   `json:"desc"`
	Values []string `json:"values"`
}
type laohuangliCache struct {
	Date string           `json:"date"`
	ID   map[int64]string `json:"caches"`
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

func (lhl *laohuangli) init(db *scribble.Driver) {
	*lhl = laohuangli{
		entries: make([]entry, 0),
	}
	db.Read("datas", "laohuangli", &lhl.entries)
	db.Read("datas", "templates", &lhl.templates)
	db.Read("datas", "cache", &lhl.cache)
	lhl.createBanlancedEntries()
}

// 计算字符串的模板实例深度之和
func (lhl *laohuangli) getTemplateDepth(s string) int {
	depth := 1
	sTmpl := fasttemplate.New(s, "{{", "}}")
	sTmpl.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		if _, ok := lhl.templates[tag]; ok {
			depth += len(lhl.templates[tag].Values)
			return w.Write([]byte(""))
		}
		depth = 0
		return 0, errors.New("invalid template")
	})
	return depth
}

// 由原始词条库生成均衡词条库
func (lhl *laohuangli) createBanlancedEntries() {
	lhl.entriesBanlanced = make([]entry, 0)
	for _, v := range lhl.entries {
		depth := lhl.getTemplateDepth(v.Content)
		if depth > 0 {
			lhl.entriesBanlanced = append(lhl.entriesBanlanced, v)
			if depth > 1 {
				for i := 0; i < int(math.Round(math.Log(float64(depth)))); i++ {
					lhl.entriesBanlanced = append(lhl.entriesBanlanced, v)
				}
			}
		}
	}
}

func (lhl *laohuangli) add(l entry) {
	lhl.entries = append(lhl.entries, l)
}
func (lhl *laohuangli) save() {
	db.Write("datas", "laohuangli", lhl.entries)
}
func (lhl *laohuangli) remove(c string) bool {
	// TODO:
	return false
}

func (lhl *laohuangli) random() (posStr string, negStr string, err error) {
	if len(lhl.entriesBanlanced) == 0 {
		return "", "", errors.New("没有词条")
	}
	max := big.NewInt(int64(len(lhl.entriesBanlanced)))
	p, _ := rand.Int(rand.Reader, max)
	n, _ := rand.Int(rand.Reader, max)
	posStr = lhl.entriesBanlanced[p.Int64()].Content
	negStr = lhl.entriesBanlanced[n.Int64()].Content

	buildStr := func(t *fasttemplate.Template) string {
		return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
			if _, ok := lhl.templates[tag]; ok {
				p, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lhl.templates[tag].Values))))
				return w.Write([]byte(lhl.templates[tag].Values[p.Int64()]))
			}
			return w.Write([]byte("`错误模板`"))
		})
	}

	if lhl.getTemplateDepth(posStr) > 0 {
		posTmpl := fasttemplate.New(posStr, "{{", "}}")
		posStr = buildStr(posTmpl)
	} else {
		return "", "", errors.New(posStr)
	}
	if lhl.getTemplateDepth(negStr) > 0 {
		negTmpl := fasttemplate.New(negStr, "{{", "}}")
		negStr = buildStr(negTmpl)
	} else {
		return "", "", errors.New(negStr)
	}

	if strutil.Similarity(posStr, negStr, gStrCompareAlgo) > 0.95 {
		if p.Cmp(n) > 0 {
			return "", "诸事不宜。请谨慎行事。", nil
		} else {
			return "诸事皆宜。愿好运与你同行。", "", nil
		}
	} else {
		return posStr, negStr, nil
	}
}

func (lhl *laohuangli) randomToday(id int64) string {
	_, exist := lhl.cache.ID[id]
	if !exist {
		p, n, err := lhl.random()
		if err != nil {
			return "发现错误模板，请上报管理员:\n" + err.Error()
		}
		if p != "" && n != "" {
			lhl.cache.ID[id] = "今日:\n宜" + p + "，忌" + n
		} else {
			lhl.cache.ID[id] = "今日:\n" + p + n
		}
		db.Write("datas", "cache", lhl.cache)
	}
	return lhl.cache.ID[id]
}
func (lhl *laohuangli) update() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		date := time.Now().In(gTimezone).Format("2006-01-02")
		if date != lhl.cache.Date {
			lhl.cache = laohuangliCache{Date: time.Now().In(gTimezone).Format("2006-01-02"), ID: make(map[int64]string)}
			db.Write("datas", "cache", lhl.cache)
		}
	}
}
func (lhl *laohuangli) start() {
	go lhl.update()
}
func (lhl *laohuangli) stop() {
	// TODO:
}

var laoHL laohuangli

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
			Title: "今日我的老黄历",
			Text:  fullName(c.Sender()) + " " + laoHL.randomToday(c.Sender().ID),
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
