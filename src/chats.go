package main

import (
	"strings"
	"time"

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

func chatLoad(id int64) privateChat {
	var chat privateChat
	chatx, _ := chats.Load(id)
	chat = chatx.(privateChat)
	return chat
}

func init() {
	chats = syncmap.Map{}
	go updateChats()
}

func MsgOnChat(c tele.Context) {
	if _, ok := chats.Load(c.Chat().ID); !ok {
		chats.Store(c.Chat().ID, privateChat{
			State: IDLE,
		})
	}
	chat := chatLoad(c.Chat().ID)
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
	chats.Store(c.Chat().ID, chat)
}

func updateChats() {
	second := time.NewTicker(10 * time.Second)
	for range second.C {
		chats.Range(func(i, v any) bool {
			chat := v.(privateChat)
			if chat.Timeout <= 0 {
				chats.Delete(i)
			} else {
				chat.Timeout--
				chats.Store(i, chat)
			}
			return true
		})
	}
}
