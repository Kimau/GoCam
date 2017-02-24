package main

import (
	"fmt"
	"image"
	"time"

	"github.com/tucnak/telebot"
)

const (
	REFRESH_TIME = 1 * time.Second
)

func checkAuth(id int) bool {
	for _, v := range AuthUserIDList {
		if v == id {
			return true
		}
	}
	return false
}

type Cambot struct {
	bot      *telebot.Bot
	cams     map[string]chan image.Image
	messages chan telebot.Message
}

func (b *Cambot) AddCamera(camName string, camFeed chan image.Image) {
	b.cams[camName] = camFeed
}

func (b *Cambot) SendUpdate() {

	uList := []*telebot.User{}

	for _, v := range AuthUserIDList {
		uList = append(uList, &telebot.User{ID: v})
	}

	for _, v := range b.cams {
		fn := "_image.jpg"
		<-v
		rgbImg := <-v
		// DrawClock(rgbImg, time.Now())
		saveJPEGToFolder(fn, rgbImg)

		photofile, _ := telebot.NewFile(fn)
		photo := telebot.Photo{File: photofile}

		for _, u := range uList {
			_ = b.bot.SendPhoto(u, &photo, nil)
		}

	}
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

		camNames := []string{}
		for k, _ := range b.cams {
			camNames = append(camNames, k)
		}

		replySendOpt.ReplyMarkup.CustomKeyboard = [][]string{{"hi"}, camNames}

		if message.Text == "hi" {
			text := fmt.Sprintf("Hello %s your id is %d", message.Sender.FirstName, message.Sender.ID)
			b.bot.SendMessage(message.Chat, text, &replySendOpt)

			continue messageLoop
		} else {
			v, ok := b.cams[message.Text]
			if ok {
				fn := "_image.jpg"
				<-v
				rgbImg := <-v
				// DrawClock(rgbImg, time.Now())
				saveJPEGToFolder(fn, rgbImg)

				photofile, _ := telebot.NewFile(fn)
				photo := telebot.Photo{File: photofile}
				_ = b.bot.SendPhoto(message.Chat, &photo, &replySendOpt)

				continue messageLoop
			}
		}

		b.bot.SendMessage(message.Chat, "Invalid Command", &replySendOpt)
	}
}

func startBot() *Cambot {
	var err error
	var cb Cambot

	cb.bot, err = telebot.NewBot(TelegramSecretToken)
	if err != nil {
		return nil
	}

	cb.cams = make(map[string]chan image.Image)
	cb.messages = make(chan telebot.Message)
	cb.bot.Listen(cb.messages, REFRESH_TIME)

	go cb.ProceessMessage()

	return &cb
}
