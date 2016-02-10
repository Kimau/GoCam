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

func startBot() {
	bot, err := telebot.NewBot(TELEGRAM_SECRET_TOKEN)
	if err != nil {
		return
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, REFRESH_TIME)

	for message := range messages {

		if !checkAuth(message.Sender.ID) {
			text := fmt.Sprintf("Sorry %s BITCH! You are not my boss. Your ID is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, nil)
			continue
		}

		if message.Text == "/hi" {
			text := fmt.Sprintf("Hello %s your id is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, nil)
		} else if message.Text == "/cam" {

			for _, v := range [...]string{"camA", "camB"} {
				filename := fmt.Sprintf("%s/%s %d.jpeg", CAPTURE_FOLDER, v, 1)
				photofile, _ := telebot.NewFile(filename)
				photo := telebot.Photo{File: photofile}
				_ = bot.SendPhoto(message.Chat, &photo, nil)
			}

		} else {
			bot.SendMessage(message.Chat, "Say /hi", nil)
		}
	}
}
