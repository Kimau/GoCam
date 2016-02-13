package main

import (
	"fmt"
	"github.com/tucnak/telebot"
	"image"
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
			CustomKeyboard:  [][]string{[]string{"/hi", "/cam", "/lum"}},
			OneTimeKeyboard: true,
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

			imgList := []image.Image{}
			for _, v := range camObjs {
				v.lock.Lock()
				imgList = append(imgList, v.lastImg)
				v.lock.Unlock()
			}

			finalImg := mergeImage(imgList)
			saveJPEGToFolder("_temp.jpg", finalImg)

			photofile, _ := telebot.NewFile("_temp.jpg")
			photo := telebot.Photo{File: photofile}
			_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)

		} else if message.Text == "/lum" {
			text := "Lum \n"
			for _, cam := range camObjs {
				text += cam.name + ": ["
				for _, v := range cam.data {
					text += fmt.Sprintf("(%d,%f)", v.lum, v.frameDuration.Seconds())
				}

				text += "] \n"
			}

			fmt.Println(text)

			bot.SendMessage(message.Chat, text, &replySendOpt)
		} else {
			bot.SendMessage(message.Chat, "Say */hi*", &replySendOpt)
		}
	}
}
