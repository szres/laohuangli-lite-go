package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/syncmap"
	tele "gopkg.in/telebot.v3"
)

const (
	IDLE int = iota
	NOMINATE
)

type privateChat struct {
	State   int
	Timeout int
}

var chats = syncmap.Map{}
var adminCMD []string
var userCMD []string
var chatCMD []string

func chatLoad(id int64) privateChat {
	var chat privateChat
	chatx, _ := chats.Load(id)
	chat = chatx.(privateChat)
	return chat
}

func init() {
	adminCMD = []string{
		"/listall", "/forcereadlocal", "/random", "/randommore",
	}
	userCMD = []string{
		"/help", "/start", "/nominate", "/list",
	}
	chatCMD = append(adminCMD, userCMD...)
	chats = syncmap.Map{}
	go updateChats()
}
func msgDelete() func(c tele.Context) error {
	return func(c tele.Context) error {
		n := nominations.pickByID(c.Message().ID)
		if n != nil {
			c.Send("提名词条 \"`"+n.Content+"`\" 已被删除，投票立即失效", tele.ModeMarkdownV2)
			nominations.remove(c.Message().ID)
		}
		defer func() {
			c.Delete()
		}()
		return nil
	}
}
func msg2User(userID int64, what any) error {
	chat, chaterr := b.ChatByID(userID)
	if chaterr == nil {
		_, err := b.Send(chat, what, tele.ParseMode(tele.ModeMarkdownV2))
		return err
	}
	return chaterr
}

func cmdInChatHandler(c tele.Context) error {
	if _, ok := chats.Load(c.Chat().ID); !ok {
		chats.Store(c.Chat().ID, privateChat{
			State: IDLE,
		})
	}
	chat := chatLoad(c.Chat().ID)
	defer func() {
		chat.Timeout = 9
		chats.Store(c.Chat().ID, chat)
	}()

	if slices.Contains(adminCMD, c.Text()) {
		if c.Sender().ID != gAdminID {
			return c.Send("您没有权限使用此命令")
		}
	}

	randLaoHuangLi := func() string {
		a, err := laoHL.randomNotDelete()
		b, _ := laoHL.randomNotDelete()
		if err != nil {
			return fmt.Sprintf("错误:\\[`%s`\\]", err.Error())
		}
		return fmt.Sprintf("宜`%s` 忌`%s`", a, b)
	}

	switch c.Text() {
	case "/start":
		fallthrough
	case "/help":
		chat.State = IDLE
		help := "提名新词条请发送 /nominate\n列举提名词条请发送 /list"
		webDomain := os.Getenv("WEB_DOMAIN")
		if webDomain != "" {
			help += "\n\n查看其他信息请访问老黄历网站: " + webDomain
		}
		if c.Sender().ID == gAdminID {
			help += "\n\n以下为管理员命令：\n列出所有提名词条请发送 /listall\n强制读取本地词条请发送 /forcereadlocal\n获取一个随机提名词条请发送 /random\n获取多个随机提名词条请发送 /randommore"
		}
		return c.Send(help)
	case "/list":
		chat.State = IDLE
		existNomination := 0
		var msg string
		for _, v := range nominations {
			if v.NominatorID == c.Sender().ID {
				existNomination++
				msg += fmt.Sprintf("提名词条 \"`%s`\" 赞成 `%d` 票，反对 `%d` 票\n", v.Content, len(v.ApprovedUsers), len(v.RefusedUsers))
			}
		}
		if existNomination == 0 {
			return c.Send("你还没有提名任何词条，请发送 /nominate 提名新词条")
		} else {
			return c.Send(msg, tele.ModeMarkdownV2)
		}
	case "/nominate":
		existNomination := 0
		for _, v := range nominations {
			if v.NominatorID == c.Sender().ID {
				existNomination++
			}
		}
		if existNomination >= 5 {
			chat.State = IDLE
			return c.Send("你已经提名过太多词条了，请等待提名投票结束再提交新词条吧！")
		}
		chat.State = NOMINATE
		return c.Send("请输入你要提名的词条内容：")

	case "/forcereadlocal":
		laoHL.init(db)
		nominations.init()
		return c.Send("已强制读取本地储存", tele.ModeMarkdownV2)
	case "/random":
		return c.Send(randLaoHuangLi(), tele.ModeMarkdownV2)
	case "/randommore":
		ret := "Result:"
		for i := 0; i < 10; i++ {
			ret += fmt.Sprintf("\n%02d: %s", i, randLaoHuangLi())
		}
		return c.Send(ret, tele.ModeMarkdownV2)
	case "/listall":
		var msg string
		for i, v := range nominations {
			msg += fmt.Sprintf("%d\\. `%s` 提名词条 \"`%s`\" 赞成 `%d` 票，反对 `%d` 票\n结束时间: `%s`\n", i+1, v.NominatorName, v.Content, len(v.ApprovedUsers), len(v.RefusedUsers), v.voteEndTimeString())
		}
		if msg == "" {
			msg = "当前没有提名任何词条"
		}
		return c.Send(msg, tele.ModeMarkdownV2)
	}
	return nil
}
func msgInChatHandler(c tele.Context) error {
	senderName := fullName(c.Sender())
	if _, ok := chats.Load(c.Chat().ID); !ok {
		chats.Store(c.Chat().ID, privateChat{
			State: IDLE,
		})
	}
	chat := chatLoad(c.Chat().ID)
	defer func() {
		chat.Timeout = 9
		chats.Store(c.Chat().ID, chat)
	}()

	switch chat.State {
	case NOMINATE:
		nominate := strings.TrimSpace(c.Text())
		if len([]rune(nominate)) < 1 {
			return c.Send("提名内容太短，请重新提名。")
		}
		if nominate[0] == '/' {
			return c.Send("格式错误，请重新提名。")
		}
		success, reason := nominationValidCheck(nominate, senderName)
		for _, v := range reason {
			c.Send(v)
		}
		chat.State = IDLE
		if success == -1 {
			chat.State = NOMINATE
		}
		if success == 0 {
			mk := &tele.ReplyMarkup{ResizeKeyboard: true}
			publishBtn := mk.Query("发布", "nominate")
			deleteBtn := mk.Data("删除", "deleteBtn")

			mk.Inline(mk.Row(publishBtn, deleteBtn))
			newNomination := nomination{
				Content:       nominate,
				CID:           c.Chat().ID,
				NominatorName: senderName,
				NominatorID:   c.Sender().ID,
				Time:          time.Now().Unix(),
				ApprovedUsers: make([]int64, 0),
				RefusedUsers:  make([]int64, 0),
			}
			msg, err := b.Send(c.Chat(), newNomination.buildVotingText(), mk, tele.ModeMarkdownV2)
			if err != nil {
				fmt.Println(err)
			} else {
				newNomination.UUID = uuid.NewV4().String()
				newNomination.ID = msg.ID
				nominations.add(newNomination)
			}
			return err
		}
	}
	return nil
}

func updateChats() {
	second := time.NewTicker(10 * time.Second)
	for range second.C {
		chats.Range(func(i, v any) bool {
			chat := v.(privateChat)
			if chat.Timeout <= 0 {
				if chatLoad(i.(int64)).State == NOMINATE {
					msg2User(i.(int64), "提名已超时，若需重新提名，请再次发送 /nominate。")
				}
				chats.Delete(i)
			} else {
				chat.Timeout--
				chats.Store(i, chat)
			}
			return true
		})
	}
}
