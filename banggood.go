package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	referallParameter string = "p"
	referalID         string = "29181220952666201804"

	bgTemplate string = `От: <a href="tg://user?id=%d">%s</a>%s
Ссылка: <a href="%s">%s</a>
`
)

type Banggood struct {
	RefID string
	bot   *tgbotapi.BotAPI
}

func InitBanggood(bot *tgbotapi.BotAPI) *Banggood {
	return &Banggood{
		RefID: referalID,
		bot:   bot,
	}
}

func (b *Banggood) Check(update tgbotapi.Update) {
	fromChat := update.Message.ForwardFromChat
	if fromChat != nil && fromChat.ID == -1001193714500 {
		return
	}

	ok, u := hasLink(update.Message.Text)
	if !ok {
		return
	}

	if !b.itsMe(u) {
		return
	}

	uri, q, err := extractURL(u)
	if err != nil {
		return
	}

	resp, err := http.Get(uri.String())
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	regularExpression := regexp.MustCompile(`<title>(.*?)<\/title>`)
	title := regularExpression.FindString(string(body))
	title = strings.TrimPrefix(title, "<title>")
	title = strings.TrimSuffix(title, "</title>")
	title = html.UnescapeString(title)

	incMsg := update.Message.Text

	othText := strings.Replace(incMsg, u, "", -1)
	if othText != "" {
		othText = fmt.Sprintf("\nСообщение: %s", othText)
	}

	if resp.Request != nil && resp.Request.Response != nil {
		loc := resp.Request.Response.Header.Get("Location")
		if loc != "" {
			uri, _ = url.Parse(loc)
		}

	}
	uri = b.replaceReferal(uri, q)
	uri.Host = "ru.banggood.com"

	uris := uri.String()

	text := fmt.Sprintf(bgTemplate,
		update.Message.From.ID, update.Message.From.String(),
		othText,
		uris, title,
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"

	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error(err)
		return
	}

	go b.deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
}

func (b *Banggood) deleteMessage(chatID int64, messageID int) error {
	_, err := b.bot.Send(tgbotapi.NewDeleteMessage(chatID, messageID))

	return err
}

func (b *Banggood) replaceReferal(u *url.URL, q url.Values) *url.URL {
	nq := make(url.Values)
	nq.Add(referallParameter, b.RefID)

	u.RawQuery = nq.Encode()

	return u
}

func (b *Banggood) itsMe(v string) bool {
	return strings.Contains(v, "banggood")
}
