package main

import (
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	IDLE int = iota
	NOMINATE
	NOMINATE_FORCE
)

type privateChat struct {
	State   int
	Timeout int
}

var chats map[int64]privateChat

func init() {
	chats = make(map[int64]privateChat)
	go updateChats()
}

func MsgOnChat(c tele.Context) {
	if _, ok := chats[c.Chat().ID]; !ok {
		chats[c.Chat().ID] = privateChat{
			State: IDLE,
		}
	}
	chat := chats[c.Chat().ID]
	chat.Timeout = 60
	switch chat.State {
	case NOMINATE_FORCE:
		_, reason := addNomination(c.Text(), c.Sender().FirstName+c.Sender().LastName, true)
		c.Send(reason)
		chat.State = IDLE
	case NOMINATE:
		res, reason := addNomination(c.Text(), c.Sender().FirstName+c.Sender().LastName, false)
		c.Send(reason)
		if res == -1 {
			c.Send("如果确认你的提名确实有效，可以发送 /fnominate 进行强制提名")
		}
		chat.State = IDLE
	case IDLE:
		if c.Text() == "/nominate" {
			chat.State = NOMINATE
		}
		if c.Text() == "/fnominate" {
			chat.State = NOMINATE_FORCE
		}
	}
	chats[c.Chat().ID] = chat
}

func updateChats() {
	second := time.NewTicker(1 * time.Second)
	for range second.C {
		for i, v := range chats {
			chat := v
			if chat.Timeout <= 0 {
				delete(chats, i)
			} else {
				chat.Timeout--
				chats[i] = chat
			}
		}
	}
}
