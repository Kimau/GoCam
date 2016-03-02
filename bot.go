package main

import (
	"fmt"
	"time"

	"github.com/tucnak/telebot"
)

const (
	REFRESH_TIME = 1 * time.Second
)

func checkAuth(id int) bool {
	for _, v := range AUTH_ID_LIST {
		if v == id {
			return true
		}
	}
	return false
}

type Cambot struct {
	bot      *telebot.Bot
	camNames []string
	camFeeds []chan string
	messages chan telebot.Message
}

func (b *Cambot) AddCamera(camName string, camFeed chan string) {
	b.camNames = append(b.camNames, camName)
	b.camFeeds = append(b.camFeeds, camFeed)
}

func (b *Cambot) ProceessMessage() {

	replySendOpt := telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply:      true,
			CustomKeyboard:  [][]string{{"/hi"}},
			OneTimeKeyboard: false,
			ResizeKeyboard:  true,
		}}

messageLoop:
	for {
		message, ok := <-b.messages
		if !ok {
			return
		}

		if !checkAuth(message.Sender.ID) {
			text := fmt.Sprintf("Sorry %s BITCH! You are not my boss. Your ID is %d", message.Sender.FirstName, message.Sender.ID)
			b.bot.SendMessage(message.Chat, text, nil)
			continue
		}

		replySendOpt.ReplyMarkup.CustomKeyboard = [][]string{{"/hi"}, b.camNames}

		if message.Text == "/hi" {
			text := fmt.Sprintf("Hello %s your id is %d", message.Sender.FirstName, message.Sender.ID)
			b.bot.SendMessage(message.Chat, text, &replySendOpt)
		}

		for i, camName := range b.camNames {
			if message.Text[1:] == camName {
				fn := <-b.camFeeds[i]
				photofile, _ := telebot.NewFile(fn)
				photo := telebot.Photo{File: photofile}
				_ = b.bot.SendPhoto(message.Chat, &photo, &replySendOpt)
				continue messageLoop
			}
		}

		b.bot.SendMessage(message.Chat, "Say */hi*", &replySendOpt)
	}
}

func startBot() *Cambot {
	var err error
	var cb Cambot

	cb.bot, err = telebot.NewBot(TELEGRAM_SECRET_TOKEN)
	if err != nil {
		return nil
	}

	cb.messages = make(chan telebot.Message)
	cb.bot.Listen(cb.messages, REFRESH_TIME)

	go cb.ProceessMessage()

	return &cb
}
