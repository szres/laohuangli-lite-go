package main

import (
	"strings"
	"time"

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
	chat.Timeout = 9
	switch chat.State {
	case NOMINATE:
		nominate := strings.TrimSpace(c.Text())
		success, reason := pushNomination(nominate, c.Sender().FirstName+c.Sender().LastName)
		for _, v := range reason {
			c.Send(v)
		}
		chat.State = IDLE
	case IDLE:
		if c.Text() == "/nominate" {
			chat.State = NOMINATE
		}
	}
	chats[c.Chat().ID] = chat
}

func updateChats() {
	second := time.NewTicker(10 * time.Second)
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
