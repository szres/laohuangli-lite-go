package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	tele "gopkg.in/telebot.v3"
)

type nomination struct {
	UUID          string  `json:"uuid"`
	ID            int     `json:"id"`
	Content       string  `json:"content"`
	NominatorName string  `json:"nominator"`
	NominatorID   int64   `json:"nominatorID"`
	Time          int64   `json:"time"`
	ApprovedUsers []int64 `json:"approvedUsers"`
	RefusedUsers  []int64 `json:"refusedUsers"`
}

var nominations []nomination
var isNominationUpdated bool = false

func init() {
	nominations = make([]nomination, 0)
	db.Read("datas", "nominations", &nominations)
	go updateNominations()
}

func updateNominations() {
	minute := time.NewTicker(1 * time.Minute)
	for range minute.C {
		isApproved := false
		newNominations := make([]nomination, 0)
		for _, v := range nominations {
			if time.Now().Unix() >= v.Time+86400 {
				chat, chaterr := b.ChatByID(v.NominatorID)
				isNominationUpdated = true
				if len(v.ApprovedUsers) >= 5 && len(v.RefusedUsers) < len(v.ApprovedUsers)/2 {
					isApproved = true
					laohuangliList = append(laohuangliList, laohuangli{
						Content:   v.Content,
						Nominator: v.NominatorName,
					})
					if chaterr == nil {
						b.Send(chat, fmt.Sprintf("恭喜你提名的词条 \"`%s`\" 最终投票结果为赞成票 `%d` 票，反对票 `%d` 票，达到上线要求。现在已经正式上线。", v.Content, len(v.ApprovedUsers), len(v.RefusedUsers)), tele.ModeMarkdownV2)
					}
				} else {
					if chaterr == nil {
						b.Send(chat, fmt.Sprintf("非常遗憾，你提名的词条 \"`%s`\" 最终投票结果为赞成票 `%d` 票，反对票 `%d` 票，未达到上线要求，无法上线。", v.Content, len(v.ApprovedUsers), len(v.RefusedUsers)), tele.ModeMarkdownV2)
					}
				}
			} else {
				newNominations = append(newNominations, v)
			}
		}
		nominations = newNominations
		if isNominationUpdated {
			saveNominations()
			isNominationUpdated = false
		}
		if isApproved {
			saveLaohuangli()
		}
	}
}

func saveNominations() {
	db.Write("datas", "nominations", &nominations)
}

func addNomination(n nomination) {
	nominations = append(nominations, n)
	saveNominations()
}

func getNomination(id int) *nomination {
	for _, v := range nominations {
		if v.ID == id {
			return &v
		}
	}
	return nil
}

func deleteNomination(id int) {
	for i, v := range nominations {
		if v.ID == id {
			nominations = append(nominations[:i], nominations[i+1:]...)
			break
		}
	}
	saveNominations()
}

type similarContent struct {
	Similarity float64
	Content    string
	Nominator  string
}

func dupNominationCheck(content string, nominator string) (result int, response []string) {
	response = make([]string, 0)
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
			Nominator:  v.NominatorName,
		})
	}
	if len(similarNominations) > 0 && similarNominations[0].Similarity > 0.9 {
		result = -1
		response = append(response, "提名内容与 "+similarNominations[0].Nominator+" 提名的 \""+similarNominations[0].Content+"\" 相似度过高，请更换提名的词条")
		return
	}

	result = 0
	if len(similarNominations) > 0 {
		resp := "提名词条与以下词条相似:\n"
		for i, v := range similarNominations {
			resp += strconv.Itoa(i+1) + ". 由 " + v.Nominator + " 提名的 \"" + v.Content + "\" - 相似度 " + strconv.FormatFloat(v.Similarity, 'f', 2, 64) + "\n"
		}
		resp += "请确认词条无重复后再发布投票。"
		response = append(response, resp)
	}
	response = append(response, "提名词条 \""+content+"\" 投票已生成，发布投票后将进入投票阶段。\n24 小时内获得不少于 5 个赞成票且反对票数不多于赞成票数的一半，词条即可上线")
	return
}

func buildVoteText(n nomination) string {
	return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已开始投票。\n请为此词条是否可以加入老黄历每日算命结果投出神圣的一票吧！\n\n赞成：`%d` 票\n反对：`%d` 票\n\n投票将于 `%s` 结束", fmt.Sprintf("[%s](tg://user?id=%d)", n.NominatorName, n.NominatorID), n.Content, len(n.ApprovedUsers), len(n.RefusedUsers), time.Unix(n.Time+86400, 0).In(gTimezone).Format("2006-01-02 15:04"))
}
func buildVoteMarkup(n nomination) *tele.ReplyMarkup {
	mk := &tele.ReplyMarkup{ResizeKeyboard: true}
	voteApproveBtn := mk.Data("赞成", "voteApproveBtn", n.UUID)
	voteRefuseBtn := mk.Data("反对", "voteRefuseBtn", n.UUID)
	mk.Inline(mk.Row(voteApproveBtn, voteRefuseBtn))
	b.Handle(&voteApproveBtn, voteApprove())
	b.Handle(&voteRefuseBtn, voteRefuse())
	return mk
}

func buildVotes(n nomination) (result *tele.ArticleResult) {
	result = &tele.ArticleResult{}
	result.Title = "发布词条 " + n.Content + " 的投票"
	result.Text = buildVoteText(n)
	result.ParseMode = tele.ModeMarkdownV2
	result.ReplyMarkup = buildVoteMarkup(n)
	return
}

func removeUserFromList(user int64, list []int64) []int64 {
	for i, v := range list {
		if v == user {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
func addUserToList(user int64, list []int64) []int64 {
	for _, v := range list {
		if v == user {
			return list
		}
	}
	return append(list, user)
}
func isUserExistInList(user int64, list []int64) bool {
	for _, v := range list {
		if v == user {
			return true
		}
	}
	return false
}
func voteApprove() func(c tele.Context) error {
	return func(c tele.Context) (err error) {
		for idx, n := range nominations {
			if n.UUID == c.Data() {
				isNominationUpdated = true
				if isUserExistInList(c.Sender().ID, n.ApprovedUsers) {
					nominations[idx].ApprovedUsers = removeUserFromList(c.Sender().ID, n.ApprovedUsers)
					c.Respond(&tele.CallbackResponse{
						Text: "您取消了赞成票",
					})
					err = c.Edit(buildVoteText(nominations[idx]), buildVoteMarkup(nominations[idx]), tele.ModeMarkdownV2)
					return
				} else {
					nominations[idx].ApprovedUsers = addUserToList(c.Sender().ID, n.ApprovedUsers)
					nominations[idx].RefusedUsers = removeUserFromList(c.Sender().ID, n.RefusedUsers)
					c.Respond(&tele.CallbackResponse{
						Text: "您投出了赞成票",
					})
					err = c.Edit(buildVoteText(nominations[idx]), buildVoteMarkup(nominations[idx]), tele.ModeMarkdownV2)
					return
				}
			}
		}
		c.Edit(c.Text(), tele.ModeMarkdownV2)
		return c.Respond(&tele.CallbackResponse{
			Text: "投票已失效",
		})
	}
}
func voteRefuse() func(c tele.Context) error {
	return func(c tele.Context) (err error) {
		for idx, n := range nominations {
			if n.UUID == c.Data() {
				isNominationUpdated = true
				if isUserExistInList(c.Sender().ID, n.RefusedUsers) {
					nominations[idx].RefusedUsers = removeUserFromList(c.Sender().ID, n.RefusedUsers)
					c.Respond(&tele.CallbackResponse{
						Text: "您取消了反对票",
					})
					err = c.Edit(buildVoteText(nominations[idx]), buildVoteMarkup(nominations[idx]), tele.ModeMarkdownV2)
					return
				} else {
					nominations[idx].RefusedUsers = addUserToList(c.Sender().ID, n.RefusedUsers)
					nominations[idx].ApprovedUsers = removeUserFromList(c.Sender().ID, n.ApprovedUsers)
					c.Respond(&tele.CallbackResponse{
						Text: "您投出了反对票",
					})
					err = c.Edit(buildVoteText(nominations[idx]), buildVoteMarkup(nominations[idx]), tele.ModeMarkdownV2)
					return
				}
			}
		}
		c.Edit(c.Text(), tele.ModeMarkdownV2)
		return c.Respond(&tele.CallbackResponse{
			Text: "投票已失效",
		})
	}
}
