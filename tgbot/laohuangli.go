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
	"gonum.org/v1/gonum/stat/combin"
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

func (lhl *laohuangli) randomEntryIndex() (idx int64, err error) {
	if len(lhl.entriesBanlanced) == 0 {
		lhl.createBanlancedEntries()
		if len(lhl.entriesBanlanced) == 0 {
			return 0, errors.New("没有词条")
		}
	}
	max := big.NewInt(int64(len(lhl.entriesBanlanced)))
	i, _ := rand.Int(rand.Reader, max)
	idx = i.Int64()
	return
}

func buildStrFromTmpl(t *fasttemplate.Template, tmpl map[string]laohuangliTemplate) string {
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if _, ok := tmpl[tag]; ok {
			p, _ := rand.Int(rand.Reader, big.NewInt(int64(len(tmpl[tag].Values))))
			return w.Write([]byte(tmpl[tag].Values[p.Int64()]))
		}
		return w.Write([]byte("`错误模板`"))
	})
}
func buildStrFromTmplWoDup(t *fasttemplate.Template, tmpl map[string]laohuangliTemplate) string {
	// 此方法会移除掉模板中选中的项，使得每个模板项只会被选择一次
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if _, ok := tmpl[tag]; ok {
			p, _ := rand.Int(rand.Reader, big.NewInt(int64(len(tmpl[tag].Values))))
			ret := tmpl[tag].Values[p.Int64()]
			temp := tmpl[tag]
			temp.Values = append(tmpl[tag].Values[:p.Int64()], tmpl[tag].Values[p.Int64()+1:]...)
			tmpl[tag] = temp
			return w.Write([]byte(ret))
		}
		return w.Write([]byte("`错误模板`"))
	})
}

func (lhl *laohuangli) randomStringAndIndex() (idx int64, str string, err error) {
	idx, err = lhl.randomEntryIndex()
	if err != nil {
		return
	}
	str = lhl.entriesBanlanced[idx].Content

	if lhl.getTemplateDepth(str) > 0 {
		posTmpl := fasttemplate.New(str, "{{", "}}")
		str = buildStrFromTmpl(posTmpl, lhl.templates)
	} else {
		err = errors.New(str)
		return
	}
	return
}
func (lhl *laohuangli) randomNotDelete() (str string, err error) {
	_, str, err = lhl.randomStringAndIndex()
	return
}
func (lhl *laohuangli) randomThenDelete() (str string, err error) {
	n, str, err := lhl.randomStringAndIndex()
	lhl.deleteBanlancedEntries([]int64{n})
	return
}

func (lhl *laohuangli) randomToday(id int64, name string) string {
	r := lhl.cache.Exist(id)
	if len(r) == 0 {
		head := ""
		body := ""
		pp := 1
		np := 1
		switch len(lhl.cache.Caches) {
		case 0:
			pp = 4
			np = 2
			head = "作为首位窥视者洞察到了清晰的命运，今日：\n"
		case 1:
			pp = 3
			np = 1
			head = "窥视的命运已开始模糊，今日：\n"
		case 2:
			pp = 1
			np = 1
			head = "从完全模糊的命运中洞察到了一丝线索，今日：\n"
		case 12:
			pp = 2
			np = 6
			head = "作为第十三位窥视者，命运的天平将为他倾斜，今日：\n"
		default:
			pp = 1
			np = 1
			head = "今日：\n"
		}
		strSlice := make([]string, 0)
		for i := 0; i < pp+np; i++ {
			str, err := lhl.randomThenDelete()
			if err != nil {
				return "发现错误，请上报管理员:\n[ERROR]" + err.Error()
			}
			strSlice = append(strSlice, str)
			if i > 0 {
				body += "，"
			}
			if i < pp {
				body += "宜"
			} else {
				body += "忌"
			}
			body += str
		}
		body += "。"
		if strutil.Similarity(strSlice[0], strSlice[1], gStrCompareAlgo) > 0.95 {
			randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(25600)))
			if randInt.Cmp(big.NewInt(12800)) >= 0 {
				body = "诸事不宜。请谨慎行事。"
			} else {
				body = "诸事皆宜。愿好运与你同行。"
			}
		}
		lhl.cache.Push(id, name, head+body)
		lhl.cache.Save()
		return head + body
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

type results struct {
	Positive string `json:"positive"`
	Negative string `json:"negative"`
}
type todayResults struct {
	Clothing results `json:"clothing"`
	Food     results `json:"food"`
	Travel   results `json:"travel"`
}
type laohuangliCache struct {
	Date   string                     `json:"date"`
	Today  todayResults               `json:"today"`
	Caches map[int64]laohuangliResult `json:"caches"`
}

func (tr todayResults) String() (output string) {
	sh := []string{
		"未分配内存中的随机比特揭示了今日的运程",
		"磁盘坏道中的损坏数据揭示了今日的运势",
		"昨天的群聊内容预示了今天的命运走向",
		"手机麦克风收集到的录音数据预测了今天的最佳策略",
	}
	randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(sh))))
	output = sh[randInt.Int64()] + "：\n\n"
	output += "👗今日穿搭👗\n宜" + tr.Clothing.Positive + "，\n忌" + tr.Clothing.Negative + "。\n\n"
	output += "🍔今日饮食🍔\n宜" + tr.Food.Positive + "，\n忌" + tr.Food.Negative + "。\n\n"
	output += "🚗今日出行🚗\n宜" + tr.Travel.Positive + "，\n忌" + tr.Travel.Negative + "。"
	return
}

func (tr *todayResults) NewRand() {
	*tr = todayResults{
		Clothing: results{
			Positive: "穿衣",
			Negative: "全裸"},
		Food: results{
			Positive: "吃饭",
			Negative: "辟谷"},
		Travel: results{
			Positive: "直立",
			Negative: "蠕动"}}

	// 衣 - 互斥特征组
	headWear := [][]string{
		{
			"{{haircolor}}色头发",
			"{{haircolor}}色{{hairstyle}}",
			"{{hairstyle}}",
		},
		{
			"{{hat}}",
			"{{color1c}}色{{hat}}",
			"{{color1c}}色帽子",
		},
	}
	bodyWear := [][]string{
		{
			"{{topwear}}",
			"{{color1c}}色上衣",
			"{{color1c}}色{{topwear}}",
		},
		{
			"{{bottomwear}}",
			"{{color1c}}色下装",
			"{{color1c}}色{{bottomwear}}",
		},
	}
	fullbodyWear := []string{
		"{{bodywear}}",
		"{{color1c}}色{{bodywear}}",
		"{{color1c}}色套装",
	}
	underWear := []string{
		"{{underwear}}",
		"{{color1c}}色{{underwear}}",
		"{{color1c}}色内衣",
	}
	legWear := []string{
		"{{socks}}",
		"{{color1c}}色{{socks}}",
		"{{color1c}}色袜子",
	}
	footWear := []string{
		"{{shoe}}",
		"{{color1c}}色{{shoe}}",
		"{{color1c}}色鞋子",
	}

	var randInt *big.Int
	// 从[]slice中随机选取n个不重复的slice n>0
	getRandomFromSliceSlice := func(slice [][]string) (ret []string) {
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
		list := combin.Combinations(len(slice), int(randInt.Int64())+1)
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(list))))
		listPick := list[randInt.Int64()]
		for _, k := range listPick {
			randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(slice[k]))))
			ret = append(ret, slice[k][randInt.Int64()])
		}
		return
	}
	getRandomOneFromSlice := func(slice []string) (ret string) {
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
		return slice[randInt.Int64()]
	}
	getRandomNFromSlice := func(slice []string) (ret []string) {
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
		list := combin.Combinations(len(slice), int(randInt.Int64())+1)
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(list))))
		listPick := list[randInt.Int64()]
		for _, k := range listPick {
			ret = append(ret, slice[k])
		}
		return
	}

	// 食 - 互斥特征组
	food := []string{
		"吃{{food}}",
		"喝{{drink}}",
		"去{{wheretoeat}}吃{{food}}",
		"去{{wheretoeat}}喝{{drink}}",
		"吃{{food}}喝{{drink}}",
		"就着{{drink}}吃{{food}}",
		"{{food}}与{{food}}同食",
		"{{drink}}与{{drink}}同饮",
	}
	// 行 - 互斥特征组
	travel := []string{
		"{{transport}}",
		"{{transport}}",
		"{{transport}}转{{transportwo}}",
	}

	wearStr := []string{}
	foodStr := []string{}
	travelStr := []string{}
	for i := 0; i < 2; i++ {
		wearList := make([]string, 0)
		wearList = append(wearList, getRandomFromSliceSlice(headWear)...)
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(25600)))
		if randInt.Cmp(big.NewInt(19200)) >= 0 {
			wearList = append(wearList, getRandomFromSliceSlice(bodyWear)...)
		} else {
			wearList = append(wearList, getRandomOneFromSlice(fullbodyWear))
		}
		wearList = append(wearList, getRandomOneFromSlice(underWear))
		wearList = append(wearList, getRandomOneFromSlice(legWear))
		wearList = append(wearList, getRandomOneFromSlice(footWear))
		wearList = getRandomNFromSlice(wearList)

		wearStr = append(wearStr, "")
		for k, v := range wearList {
			conc := ""
			if k == 1 {
				conc = "配"
			}
			if k > 1 {
				if k == len(wearList)-1 {
					conc = "和"
				} else {
					conc = "、"
				}
			}
			wearStr[i] += conc + v
		}
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(food))))
		foodStr = append(foodStr, food[randInt.Int64()])
		randInt, _ = rand.Int(rand.Reader, big.NewInt(int64(len(travel))))
		travelStr = append(travelStr, travel[randInt.Int64()])
		if laoHL.getTemplateDepth(wearStr[i]) <= 1 || laoHL.getTemplateDepth(foodStr[i]) <= 1 || laoHL.getTemplateDepth(travelStr[i]) <= 1 {
			// 无法生成今日指引，直接返回
			return
		}
	}
	tmpl := make(map[string]laohuangliTemplate)
	for k, v := range laoHL.templates {
		tmpl[k] = v
	}
	wearStrPos := buildStrFromTmplWoDup(fasttemplate.New(wearStr[0], "{{", "}}"), tmpl)
	foodStrPos := buildStrFromTmplWoDup(fasttemplate.New(foodStr[0], "{{", "}}"), tmpl)
	travelStrPos := buildStrFromTmplWoDup(fasttemplate.New(travelStr[0], "{{", "}}"), tmpl)
	wearStrNeg := buildStrFromTmplWoDup(fasttemplate.New(wearStr[1], "{{", "}}"), tmpl)
	foodStrNeg := buildStrFromTmplWoDup(fasttemplate.New(foodStr[1], "{{", "}}"), tmpl)
	travelStrNeg := buildStrFromTmplWoDup(fasttemplate.New(travelStr[1], "{{", "}}"), tmpl)
	*tr = todayResults{
		Clothing: results{
			Positive: wearStrPos,
			Negative: wearStrNeg},
		Food: results{
			Positive: foodStrPos,
			Negative: foodStrNeg},
		Travel: results{
			Positive: travelStrPos,
			Negative: travelStrNeg},
	}
}
func (c *laohuangliCache) Init() {
	db.Read("datas", "cache", c)
}
func (c *laohuangliCache) New() {
	*c = laohuangliCache{Date: time.Now().In(gTimezone).Format("2006-01-02"), Caches: make(map[int64]laohuangliResult), Today: todayResults{}}
	c.Today.NewRand()
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
