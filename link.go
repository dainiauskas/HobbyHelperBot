package main

import (
	"fmt"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"mvdan.cc/xurls"
)

const (
	linkTemplate string = `От: <a href="tg://user?id=%d">%s</a>%s
Ссылка: <a href="%s">%s</a>
`
)

type LinkService interface {
	Check(tgbotapi.Message)
}

type LinkOp struct {
	Name     string
	RefID    string
	RefParam string

	msg         tgbotapi.Message
	url         *url.URL
	urlStr      string
	excludeFrom []int64

	api *tgbotapi.BotAPI
}

func (l LinkOp) Check(msg tgbotapi.Message) {
	l.url = nil
	l.msg = msg

	log.Debug(l.RefID, l.RefParam)

	if !l.valid() {
		return
	}

	log.Infof("Posting link of [%s]", l.Name)

	err := l.postMessage()
	if err == nil {
		go l.deleteMessage()
	}
}

func (l *LinkOp) valid() (ok bool) {
	if !l.setURL() || !l.exclude() {
		return
	}

	return true
}

func (l *LinkOp) exclude() bool {
	fw := l.msg.ForwardFromChat

	if fw != nil {
		for _, id := range l.excludeFrom {
			if fw.ID == id {
				return false
			}
		}
	}

	return true
}

func (l *LinkOp) setURL() (ok bool) {
	rxRelaxed := xurls.Relaxed
	u := rxRelaxed.FindString(l.msg.Text)
	if u == "" {
		log.Debug("Not found url in message")
		return
	}

	if !strings.Contains(u, l.Name) {
		log.Debugf("not contains name: %s", l.Name)
		return
	}

	l.urlStr = u

	if err := l.parseURL(u); err != nil {
		log.Error(err)

		return
	}

	log.Debugf("Valid host: %s", l.url.Host)

	l.setRef()

	return true
}

func (l *LinkOp) parseURL(v string) (err error) {
	log.Debug("parse value: ", v)

	l.url, err = url.Parse(v)
	if err != nil {
		return err
	}

	return err
}

func (l *LinkOp) setRef() {
	nq := make(url.Values)
	nq.Add(l.RefParam, l.RefID)

	l.url.RawQuery = nq.Encode()

	log.Debug("Ref URL: ", l.url.String())
}

func (l *LinkOp) URL() string {
	return l.url.String()
}

func (l *LinkOp) formatMessage() string {
	msg := l.msg.Text

	userMessage := strings.Replace(msg, l.urlStr, "", -1)
	if userMessage != "" {
		userMessage = fmt.Sprintf("\nСообщение: %s", userMessage)
	}

	return fmt.Sprintf(linkTemplate,
		l.msg.From.ID, l.msg.From.String(),
		userMessage,
		l.url.String(), l.URL(),
	)
}

func (l *LinkOp) postMessage() error {
	log.Debug("Creating message")

	msg := tgbotapi.NewMessage(l.msg.Chat.ID, l.formatMessage())
	msg.ParseMode = "HTML"

	log.Debug("Sending message")
	_, err := l.api.Send(msg)

	return err
}

func (l *LinkOp) deleteMessage() {
	id := l.msg.MessageID

	log.Debugf("Deleting message: %d", id)
	_, err := l.api.Send(tgbotapi.NewDeleteMessage(l.msg.Chat.ID, id))

	if err != nil {
		log.Error(err)
	}
}
