package main

import (
	"fmt"
	"github.com/tucnak/telebot"
	"time"
)

const (
	REFRESH_TIME = 1 * time.Second
)

var ()

func checkAuth(id int) bool {
	for _, v := range AUTH_ID_LIST {
		if v == id {
			return true
		}
	}
	return false
}

func startBot(camObjs []*camObject) {
	bot, err := telebot.NewBot(TELEGRAM_SECRET_TOKEN)
	if err != nil {
		return
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, REFRESH_TIME)

	replySendOpt := telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply:      true,
			CustomKeyboard:  [][]string{[]string{"/hi", "/cam"}},
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}}

	for message := range messages {

		if !checkAuth(message.Sender.ID) {
			text := fmt.Sprintf("Sorry %s BITCH! You are not my boss. Your ID is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, nil)
			continue
		}

		if message.Text == "/hi" {
			text := fmt.Sprintf("Hello %s your id is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, &replySendOpt)
		} else if message.Text == "/cam" {

			for _, v := range camObjs {
				v.lock.Lock()

				photofile, _ := telebot.NewFile(fmt.Sprintf("%s/%s.jpeg", v.folder, v.name))
				photo := telebot.Photo{File: photofile}
				_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)

				v.lock.Unlock()
			}

		} else {
			bot.SendMessage(message.Chat, "Say */hi*", &replySendOpt)
		}
	}
}
