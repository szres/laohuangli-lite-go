package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/adrg/strutil"
	tele "gopkg.in/telebot.v3"
)

type nomination struct {
	UUID          string  `json:"uuid"`
	ID            int     `json:"id"`
	CID           int64   `json:"cid"`
	LastVoteID    int     `json:"lastvoteid"`
	Content       string  `json:"content"`
	NominatorName string  `json:"nominator"`
	NominatorID   int64   `json:"nominatorID"`
	Time          int64   `json:"time"`
	ApprovedUsers []int64 `json:"approvedUsers"`
	RefusedUsers  []int64 `json:"refusedUsers"`
}
type nominationSlice []nomination

func (n *nomination) approvedBy(user int64) int {
	nominations.lazySave()
	n.RefusedUsers = removeUserFromList(user, n.RefusedUsers)
	if isUserExistInList(user, n.ApprovedUsers) {
		n.ApprovedUsers = removeUserFromList(user, n.ApprovedUsers)
		return 0
	}
	n.ApprovedUsers = addUserToList(user, n.ApprovedUsers)
	return 1
}
func (n *nomination) refusedBy(user int64) int {
	nominations.lazySave()
	n.ApprovedUsers = removeUserFromList(user, n.ApprovedUsers)
	if isUserExistInList(user, n.RefusedUsers) {
		n.RefusedUsers = removeUserFromList(user, n.RefusedUsers)
		return 0
	}
	n.RefusedUsers = addUserToList(user, n.RefusedUsers)
	return 1
}
func (n *nomination) isPassed() bool {
	if len(n.ApprovedUsers) >= 5 && len(n.RefusedUsers) < len(n.ApprovedUsers)/2 {
		return true
	}
	return false
}
func (n *nomination) isQuickPassed() bool {
	if len(n.ApprovedUsers)+len(n.RefusedUsers) >= 10 && len(n.RefusedUsers) < len(n.ApprovedUsers)/3 {
		return true
	}
	return false
}
func (n *nomination) isQuickRefused() bool {
	if len(n.ApprovedUsers)+len(n.RefusedUsers) >= 10 && len(n.RefusedUsers) > len(n.ApprovedUsers) {
		return true
	}
	return false
}
func (n *nomination) voteEndTimeString() string {
	return time.Unix(n.Time+86400, 0).In(gTimezone).Format(gTimeFormat)
}
func (n nomination) MessageSig() (messageID string, chatID int64) {
	return strconv.Itoa(n.ID), n.CID
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

var nominations nominationSlice

func (ns *nominationSlice) init() {
	*ns = make(nominationSlice, 0)
	db.Read("datas", "nominations", ns)
}
func (ns *nominationSlice) add(n nomination) {
	*ns = append(*ns, n)
	ns.save()
}
func (ns *nominationSlice) remove(id int) {
	for i, n := range *ns {
		if n.ID == id {
			*ns = append((*ns)[:i], (*ns)[i+1:]...)
			break
		}
	}
	ns.save()
}

func (ns *nominationSlice) pickByID(id int) *nomination {
	for _, v := range *ns {
		if v.ID == id {
			return &v
		}
	}
	return nil
}

func (ns *nominationSlice) save() {
	db.Write("datas", "nominations", ns)
}

var isNominationUpdated bool = false

func (ns nominationSlice) lazySave() {
	isNominationUpdated = true
}
func (ns *nominationSlice) saveRoutine() {
	if isNominationUpdated {
		ns.save()
		isNominationUpdated = false
	}
}

func (ns *nominationSlice) update() {
	minute := time.NewTicker(60 * time.Second)
	for range minute.C {
		newNominations := make(nominationSlice, 0)
		for _, v := range *ns {
			if time.Now().Unix() < v.Time+86400 && !v.isQuickPassed() && !v.isQuickRefused() {
				newNominations = append(newNominations, v)
			} else {
				if time.Now().Unix() >= v.Time+86400 && v.isPassed() || v.isQuickPassed() {
					laohuangliList.add(laohuangli{UUID: v.UUID, Content: v.Content, Nominator: v.NominatorName})
				}
				msg2User(v.NominatorID, v.buildResultMsgText())
				// TODO: Should edit last vote msg there
				b.Edit(v, v.buildVoteResultText(), tele.ModeMarkdownV2)
				ns.lazySave()
			}
		}
		*ns = newNominations
		ns.saveRoutine()
	}
}

func init() {
	nominations.init()
	go nominations.update()
}

type similarContent struct {
	Similarity float64
	Content    string
	Nominator  string
}

func nominationValidCheck(content string, nominator string) (result int, response []string) {
	response = make([]string, 0)
	if len(content) > 64 {
		result = -1
		response = append(response, "提名内容过长，请控制在 64 个字以内")
		return
	}
	if !templateValid(content) {
		result = -1
		response = append(response, "错误的模板格式或者不存在的模板变量，请检查更正后重新提交。")
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
	for _, v := range laohuangliList {
		similarity := strutil.Similarity(content, v.Content, gStrCompareAlgo)
		similarPush(similarContent{
			Similarity: similarity,
			Content:    v.Content,
			Nominator:  v.Nominator,
		})
	}
	for _, v := range nominations {
		similarity := strutil.Similarity(content, v.Content, gStrCompareAlgo)
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
	response = append(response, "提名词条 \""+content+"\" 投票已生成，发布投票后将进入投票阶段。\n获得不少于 5 个赞成票且赞成率超过 66% 的提名在24小时后可上线。\n获得不少于 10 个赞成票且赞成率超过 75% 的词条可立即上线。")
	return
}

func (n nomination) buildVoteMarkup() *tele.ReplyMarkup {
	mk := &tele.ReplyMarkup{ResizeKeyboard: true}
	voteApproveBtn := mk.Data("赞成", "voteApproveBtn", n.UUID)
	voteRefuseBtn := mk.Data("反对", "voteRefuseBtn", n.UUID)
	mk.Inline(mk.Row(voteApproveBtn, voteRefuseBtn))
	b.Handle(&voteApproveBtn, voteApprove())
	b.Handle(&voteRefuseBtn, voteRefuse())
	return mk
}

func (n nomination) buildVotingText() string {
	return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已开始投票。\n请为此词条是否可以加入老黄历每日算命结果投出神圣的一票吧！\n\n赞成：`%d` 票\n反对：`%d` 票\n\n投票将于 `%s` 结束", fmt.Sprintf("[%s](tg://user?id=%d)", n.NominatorName, n.NominatorID), n.Content, len(n.ApprovedUsers), len(n.RefusedUsers), n.voteEndTimeString())
}

func (n nomination) buildVoteResultText() string {
	if n.isQuickPassed() {
		return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已达到快速过审标准。\n\n赞成：`%d` 票\n反对：`%d` 票\n\n词条已于 `%s` 上线。", fmt.Sprintf("[%s](tg://user?id=%d)", n.NominatorName, n.NominatorID), n.Content, len(n.ApprovedUsers), len(n.RefusedUsers), time.Now().In(gTimezone).Format(gTimeFormat))
	}
	if n.isQuickRefused() {
		return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已达到快速否决条件。\n\n赞成：`%d` 票\n反对：`%d` 票\n\n词条已被系统否决。", fmt.Sprintf("[%s](tg://user?id=%d)", n.NominatorName, n.NominatorID), n.Content, len(n.ApprovedUsers), len(n.RefusedUsers))
	}
	if n.isPassed() {
		return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已达到过审标准。\n\n赞成：`%d` 票\n反对：`%d` 票\n\n词条已于 `%s` 上线。", fmt.Sprintf("[%s](tg://user?id=%d)", n.NominatorName, n.NominatorID), n.Content, len(n.ApprovedUsers), len(n.RefusedUsers), time.Now().In(gTimezone).Format(gTimeFormat))
	}
	return ""
}

func (n nomination) buildResultMsgText() string {
	if n.isQuickPassed() {
		return fmt.Sprintf("恭喜你提名的词条 \"`%s`\" 获得赞成票 `%d` 票，反对票 `%d` 票，达到快速上线要求。现在已经正式上线。", n.Content, len(n.ApprovedUsers), len(n.RefusedUsers))
	}
	if n.isQuickRefused() {
		return fmt.Sprintf("非常遗憾，你提名的词条 \"`%s`\" 获得赞成票 `%d` 票，反对票 `%d` 票，达到快速否决条件，已经被系统拒绝。", n.Content, len(n.ApprovedUsers), len(n.RefusedUsers))
	}
	if n.isPassed() {
		return fmt.Sprintf("恭喜你提名的词条 \"`%s`\" 最终投票结果为赞成票 `%d` 票，反对票 `%d` 票，达到上线要求。现在已经正式上线。", n.Content, len(n.ApprovedUsers), len(n.RefusedUsers))
	}
	return fmt.Sprintf("非常遗憾，你提名的词条 \"`%s`\" 最终投票结果为赞成票 `%d` 票，反对票 `%d` 票，未达到上线要求，无法上线。", n.Content, len(n.ApprovedUsers), len(n.RefusedUsers))
}

func buildVotes(n nomination) (result *tele.ArticleResult) {
	result = &tele.ArticleResult{}
	result.Title = "发布词条 " + n.Content + " 的投票"
	result.Text = n.buildVotingText()
	result.ParseMode = tele.ModeMarkdownV2

	mk := &tele.ReplyMarkup{ResizeKeyboard: true}
	voteApproveBtn := mk.Data("赞成", "voteApproveBtn", n.UUID)
	voteRefuseBtn := mk.Data("反对", "voteRefuseBtn", n.UUID)
	mk.Inline(mk.Row(voteApproveBtn, voteRefuseBtn))

	result.ReplyMarkup = mk
	return
}

func buildVoteResultSimple(uuid string) string {
	if uuid == "" {
		return "投票已结束"
	}
	for _, v := range laohuangliList {
		if v.UUID == uuid {
			return fmt.Sprintf("由 %s 提名的新词条 \"`%s`\" 已经通过投票正式上线。", v.Nominator, v.Content)
		}
	}
	return "投票已结束"
}

func voteApprove() func(c tele.Context) error {
	return func(c tele.Context) (err error) {
		for idx, n := range nominations {
			if n.UUID == c.Data() {
				res := nominations[idx].approvedBy(c.Sender().ID)
				if res == 0 {
					c.Respond(&tele.CallbackResponse{
						Text: "您取消了赞成票",
					})
					err = c.Edit(nominations[idx].buildVotingText(), nominations[idx].buildVoteMarkup(), tele.ModeMarkdownV2)
					return
				} else {
					c.Respond(&tele.CallbackResponse{
						Text: "您投出了赞成票",
					})
					err = c.Edit(nominations[idx].buildVotingText(), nominations[idx].buildVoteMarkup(), tele.ModeMarkdownV2)
					return
				}
			}
		}
		// TODO: show vote result better
		c.Edit(buildVoteResultSimple(c.Data()), tele.ModeMarkdownV2)
		return c.Respond(&tele.CallbackResponse{
			Text: "投票已结束",
		})
	}
}
func voteRefuse() func(c tele.Context) error {
	return func(c tele.Context) (err error) {
		for idx, n := range nominations {
			if n.UUID == c.Data() {
				res := nominations[idx].refusedBy(c.Sender().ID)
				if res == 0 {
					c.Respond(&tele.CallbackResponse{
						Text: "您取消了反对票",
					})
					err = c.Edit(nominations[idx].buildVotingText(), nominations[idx].buildVoteMarkup(), tele.ModeMarkdownV2)
					return
				} else {
					c.Respond(&tele.CallbackResponse{
						Text: "您投出了反对票",
					})
					err = c.Edit(nominations[idx].buildVotingText(), nominations[idx].buildVoteMarkup(), tele.ModeMarkdownV2)
					return
				}
			}
		}
		// TODO: show vote result better
		c.Edit(buildVoteResultSimple(c.Data()), tele.ModeMarkdownV2)
		return c.Respond(&tele.CallbackResponse{
			Text: "投票已结束",
		})
	}
}
