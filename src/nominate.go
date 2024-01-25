package main

import (
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

func addNomination(content string, nominator string, skipCheck bool) (result int, reason string) {
	if !skipCheck {
		var maxSimilarity float64 = 0
		var maxSimilarityContent string
		var maxSimilarityNominator string
		for _, v := range laohuangliList {
			similarity := strutil.Similarity(content, v.Content, metrics.NewLevenshtein())
			if similarity > maxSimilarity {
				maxSimilarity = similarity
				maxSimilarityContent = v.Content
				maxSimilarityNominator = v.Nominator
			}
		}
		for _, v := range nominations {
			similarity := strutil.Similarity(content, v.Content, metrics.NewLevenshtein())
			if similarity > maxSimilarity {
				maxSimilarity = similarity
				maxSimilarityContent = v.Content
				maxSimilarityNominator = v.Nominator
			}
		}
		if maxSimilarity > 0.7 {
			result = -1
			reason = "提名内容与 " + maxSimilarityNominator + " 提名的 \"" + maxSimilarityContent + "\" 相似度过高"
			return
		}
	}

	nominations = append(nominations, nomination{
		Content:   content,
		Nominator: nominator,
		Time:      time.Now().Unix(),
		Approved:  0,
		Refused:   0,
	})
	saveNominations()
	result = 0
	if skipCheck {
		reason = "强制"
	} else {
		reason = ""
	}
	reason += "提名词条 \"" + content + "\" 成功，进入投票阶段。\n24 小时内获得不少于 5 个赞成票且反对票数不多于赞成票数的一半，即可通过"
	return
}
