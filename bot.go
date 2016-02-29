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

func startBot(camObjs []*camObject) {
	bot, err := telebot.NewBot(TELEGRAM_SECRET_TOKEN)
	if err != nil {
		return
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, REFRESH_TIME)

	time.Sleep(1 * time.Second)

	replySendOpt := telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply:      true,
			CustomKeyboard:  [][]string{{"/cam", "/lum"}},
			OneTimeKeyboard: false,
			ResizeKeyboard:  true,
		}}

	for message := range messages {

		if !checkAuth(message.Sender.ID) {
			text := fmt.Sprintf("Sorry %s BITCH! You are not my boss. Your ID is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, &replySendOpt)
			continue
		}

		if message.Text == "/hi" {
			text := fmt.Sprintf("Hello %s your id is %d", message.Sender.FirstName, message.Sender.ID)
			bot.SendMessage(message.Chat, text, &replySendOpt)
		} else if message.Text == "/cam" {

			saveJPEGToFolder("_temp.jpg", mergeCamFeeds(camObjs))

			photofile, _ := telebot.NewFile("_temp.jpg")
			photo := telebot.Photo{File: photofile}
			_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)
		} else if message.Text == "/gif" {
			bot.SendMessage(message.Chat, "GIF can take a second", &replySendOpt)

			for _, cam := range camObjs {
				filename := fmt.Sprintf("_moving%s.gif", cam.name)
				saveAllGIFToFolder(filename, makeCamGIF(cam))
				photofile, _ := telebot.NewFile(filename)
				photo := telebot.Photo{File: photofile}
				_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)
			}

		} else if message.Text == "/lum" {

			saveGIFToFolder("_temp.gif", makeLumTimeline(camObjs), 255)

			photofile, _ := telebot.NewFile("_temp.gif")
			photo := telebot.Photo{File: photofile}
			_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)

		} else {
			bot.SendMessage(message.Chat, "Say */hi*", &replySendOpt)
		}
	}
}
