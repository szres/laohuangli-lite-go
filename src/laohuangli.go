package main

import (
	"crypto/rand"
	"errors"
	"io"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/adrg/strutil"
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/valyala/fasttemplate"
)

type laohuangli struct {
	// 本地词条
	entries []entry
	// 用户提名词条
	entriesUser []entry
	// 频次均衡后的词条
	entriesBanlanced []entry
	templates        map[string]laohuangliTemplate
	cache            laohuangliCache
}

var laoHL laohuangli

func (lhl *laohuangli) init(db *scribble.Driver) {
	*lhl = laohuangli{
		entries: make([]entry, 0),
	}
	db.Read("datas", "laohuangli", &lhl.entries)
	db.Read("datas", "templates", &lhl.templates)
	db.Read("datas", "laohuangli-user", &lhl.entriesUser)
	lhl.cache.Init()
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

	// 用户提名词条2倍权重
	for i := 0; i < 2; i++ {
		lhl.entriesBanlanced = append(lhl.entriesBanlanced, lhl.entriesUser...)
	}

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
func (lhl *laohuangli) pushBanlancedEntries(e entry) {
	// 新词条当日7倍权重
	for i := 0; i < 7; i++ {
		lhl.entriesBanlanced = append(lhl.entriesBanlanced, e)
	}
}
func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// 均衡词条库移除
func (lhl *laohuangli) deleteBanlancedEntries(s []int64) {
	s = removeDuplicate(s)
	sort.Slice(s, func(i, j int) bool {
		return s[i] > s[j]
	})

	for _, v := range s {
		if v >= int64(len(lhl.entriesBanlanced)) {
			continue
		}
		lhl.entriesBanlanced = append(lhl.entriesBanlanced[:v], lhl.entriesBanlanced[v+1:]...)
	}
}

func (lhl *laohuangli) add(l entry) {
	lhl.entriesUser = append(lhl.entriesUser, l)
}
func (lhl *laohuangli) save() {
	db.Write("datas", "laohuangli-user", lhl.entriesUser)
}
func (lhl *laohuangli) remove(c string) bool {
	// TODO:
	return false
}

func (lhl *laohuangli) random() (posStr string, negStr string, err error) {
	if len(lhl.entriesBanlanced) == 0 {
		lhl.createBanlancedEntries()
		if len(lhl.entriesBanlanced) == 0 {
			return "", "", errors.New("没有词条")
		}
	}
	max := big.NewInt(int64(len(lhl.entriesBanlanced)))
	p, _ := rand.Int(rand.Reader, max)
	n, _ := rand.Int(rand.Reader, max)
	posStr = lhl.entriesBanlanced[p.Int64()].Content
	negStr = lhl.entriesBanlanced[n.Int64()].Content
	lhl.deleteBanlancedEntries([]int64{p.Int64(), n.Int64()})

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

func (lhl *laohuangli) randomToday(id int64, name string) string {
	r := lhl.cache.Exist(id)
	if len(r) == 0 {
		p, n, err := lhl.random()
		if err != nil {
			return "发现错误模板，请上报管理员:\n" + err.Error()
		}
		if p != "" && n != "" {
			r = "今日:\n宜" + p + "，忌" + n
		} else {
			r = "今日:\n" + p + n
		}
		lhl.cache.Push(id, name, r)
		lhl.cache.Save()
	}
	return r
}
func (lhl *laohuangli) update() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		date := time.Now().In(gTimezone).Format("2006-01-02")
		if date != lhl.cache.Date {
			if len(lhl.cache.Caches) > 0 {
				lhl.cache.Backup(lhl.cache.Date)
			}
			lhl.cache.New()
			lhl.cache.Save()
			lhl.createBanlancedEntries()
		}
	}
}
func (lhl *laohuangli) start() {
	go lhl.update()
}
func (lhl *laohuangli) stop() {
	// TODO:
}

type laohuangliCache struct {
	Date   string                     `json:"date"`
	Caches map[int64]laohuangliResult `json:"caches"`
}

func (c *laohuangliCache) Init() {
	db.Read("datas", "cache", c)
}
func (c *laohuangliCache) New() {
	*c = laohuangliCache{Date: time.Now().In(gTimezone).Format("2006-01-02"), Caches: make(map[int64]laohuangliResult)}
}
func (c *laohuangliCache) Save() {
	db.Write("datas", "cache", c)
}
func (c *laohuangliCache) Backup(date string) {
	db.Write("history", date, c)
}
func (c *laohuangliCache) Exist(id int64) string {
	_, exist := c.Caches[id]
	if exist {
		return c.Caches[id].Result
	}
	return ""
}
func (c *laohuangliCache) Push(id int64, name string, content string) {
	result := laohuangliResult{Name: name, Result: content}
	c.Caches[id] = result
}
