package main

import (
	"sort"
	"strconv"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type nomination struct {
	Content   string `json:"content"`
	Nominator string `json:"nominator"`
	Time      int64  `json:"time"`
	Approved  int    `json:"approved"`
	Refused   int    `json:"refused"`
}

var nominations []nomination

func init() {
	nominations = make([]nomination, 0)
	db.Read("datas", "nominations", &nominations)
	go updateNominations()
}

func updateNominations() {
	minute := time.NewTicker(1 * time.Minute)
	for range minute.C {
		isUpdated := false
		isApproved := false
		for _, v := range nominations {
			if time.Now().Unix() >= v.Time+86400 {
				isUpdated = true
				if v.Approved >= 5 && v.Refused < v.Approved/2 {
					isApproved = true
					laohuangliList = append(laohuangliList, laohuangli{
						Content:   v.Content,
						Nominator: v.Nominator,
					})
				}
				nominations = append(nominations[:v.Refused], nominations[v.Refused+1:]...)
			}
		}
		if isUpdated {
			saveNominations()
		}
		if isApproved {
			saveLaohuangli()
		}
	}
}

func saveNominations() {
	db.Write("datas", "nominations", &nominations)
}

type similarContent struct {
	Similarity float64
	Content    string
	Nominator  string
}

func pushNomination(content string, nominator string) (result int, response []string) {
	response = make([]string, 0)
	if len(content) < 1 {
		result = -2
		response = append(response, "提名内容太短")
		return
	}
	similarNominations := make([]similarContent, 0)
	similarSort := func() {
		sort.Slice(similarNominations, func(i, j int) bool {
			return similarNominations[i].Similarity > similarNominations[j].Similarity
		})
	}
	similarPush := func(v similarContent) {
		if v.Similarity > 0.5 {
			similarNominations = append(similarNominations, v)
			similarSort()
			if len(similarNominations) > 3 {
				similarNominations = similarNominations[:3]
			}
		}
	}
	compareMethod := metrics.NewJaro()
	compareMethod.CaseSensitive = false
	for _, v := range laohuangliList {
		similarity := strutil.Similarity(content, v.Content, compareMethod)
		similarPush(similarContent{
			Similarity: similarity,
			Content:    v.Content,
			Nominator:  v.Nominator,
		})
	}
	for _, v := range nominations {
		similarity := strutil.Similarity(content, v.Content, compareMethod)
		similarPush(similarContent{
			Similarity: similarity,
			Content:    v.Content,
			Nominator:  v.Nominator,
		})
	}
	if len(similarNominations) > 0 && similarNominations[0].Similarity > 0.9 {
		result = -1
		response = append(response, "提名内容与 "+similarNominations[0].Nominator+" 提名的 \""+similarNominations[0].Content+"\" 相似度过高")
		return
	}

	// nominations = append(nominations, nomination{
	// 	Content:   content,
	// 	Nominator: nominator,
	// 	Time:      time.Now().Unix(),
	// 	Approved:  0,
	// 	Refused:   0,
	// })
	// saveNominations()
	result = 0
	if len(similarNominations) > 0 {
		resp := "提名词条与以下词条相似:\n"
		for i, v := range similarNominations {
			resp += strconv.Itoa(i+1) + ". 由 " + v.Nominator + " 提名的 \"" + v.Content + "\" - 相似度 " + strconv.FormatFloat(v.Similarity, 'f', 2, 64) + "\n"
		}
		resp += "请确认词条无重复后再发布投票。"
		response = append(response, resp)
	}
	response = append(response, "提名词条 \""+content+"\" 投票已生成(并没有)，发布投票后将进入投票阶段。\n24 小时内获得不少于 5 个赞成票且反对票数不多于赞成票数的一半，词条即可上线")
	response = append(response, "投票功能开发中，敬请期待...")
	return
}
