package main

import (
	"fmt"
	"time"

	"github.com/tucnak/telebot"
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
			CustomKeyboard:  [][]string{[]string{"/cam", "/lum"}},
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
			/*
				// Run once for each camera
				for _, v := range camObjs {
					v.lock.Lock()
					gifData := gif.GIF{
						LoopCount: -1,
					}

					numImg := len(v.imgBuffer)

					palImages := make([]*image.Paletted, numImg, numImg)
					gifData.Disposal = make([]byte, numImg, numImg)
					gifData.Delay = make([]int, numImg, numImg)

					for i, img := range v.imgBuffer {
						newPal := getColours(img)
						newPalImg := image.NewPaletted(img.Bounds(), newPal)
						draw.Draw(newPalImg, img.Bounds(), img, image.ZP, draw.Over)

						saveGIFToFolder(fmt.Sprintf("_%s_%d.gif", v.name, i), newPalImg)

						palImages = append(palImages, newPalImg)
						gifData.Disposal[i] = gif.DisposalBackground
						gifData.Delay[i] = 500
					}
					v.lock.Unlock()

					gifData.Image = palImages

					filename := fmt.Sprintf("_%s.gif", v.name)
					saveAllGIFToFolder(filename, &gifData)

					photofile, _ := telebot.NewFile(filename)
					photo := telebot.Photo{File: photofile}
					_ = bot.SendPhoto(message.Chat, &photo, &replySendOpt)
				}
				// end
			*/
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
