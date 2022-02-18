package main

import (
	"fmt"
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

	log.Infof("Authorized on account %s", bot.Self.UserName)

	exclude := []int64{-1001193714500}

	checker := []LinkService{
		LinkOp{
			Name:        "banggood",
			RefID:       os.Getenv("BANGGOOD_REF_ID"),
			RefParam:    "p",
			api:         bot,
			excludeFrom: exclude,
		},
		LinkOp{
			Name:        "radiomasterrc",
			RefID:       os.Getenv("RADIOMASTER_REF_ID"),
			RefParam:    "sca_ref",
			api:         bot,
			excludeFrom: exclude,
		},
		LinkOp{
			Name:        "betafpv",
			RefID:       os.Getenv("BETAFPV_REF_ID"),
			RefParam:    "sca_ref",
			api:         bot,
			excludeFrom: exclude,
		},
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Info("Clearing updates")
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	log.Info("Start listining")

	for update := range updates {
		log.Debug("Got update")

		if update.Message == nil {
			log.Debug("Message empty", update)
			continue
		}

		// Skipping bots
		if update.Message.From.IsBot {
			log.Debug("It is bot mssage")
			continue
		}

		fmt.Printf("%+v\n", update)

		for _, ch := range checker {
			go ch.Check(*update.Message)
		}
	}
}
