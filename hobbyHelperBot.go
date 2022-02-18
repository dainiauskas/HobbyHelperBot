package main

import (
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	Version string
	Build   string
	Name    string
)

const (
	linkTemplate string = `От: <a href="tg://user?id=%d">%s</a>%s
Ссылка: <a href="%s">%s</a>
`
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	log.Info(Name, Version, Build)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG_BOT") == "true" {
		bot.Debug = true
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Authorized on account %s", bot.Self.UserName)

	bg := InitBanggood(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Info("Clearing updates")
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	log.Info("Start listining")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Skipping bots
		if update.Message.From.IsBot {
			continue
		}

		go bg.Check(update)
	}
}
