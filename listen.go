package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/kurrik/witgo/v1/witgo"
	"github.com/juju/errors"
	"github.com/Sirupsen/logrus"
	"strconv"
	"fmt"
	"time"
)

type Bridge struct {
	tgBotAPI *tgbotapi.BotAPI
	witClient *witgo.Client
	actionClient ActionClient
}


func NewBridge(tgKey string, witKey string, actionClient ActionClient) (*Bridge, error) {
	bot, err := tgbotapi.NewBotAPI(tgKey)
	if err != nil {
		return nil, errors.Annotate(err, "Failed to connect to Telegram")
	}

	logrus.Debugf("Authorized on telegram account %s", bot.Self.UserName)

	witClient := witgo.NewClient(witKey)

	return &Bridge{
		tgBotAPI: bot,
		witClient: witClient,
		actionClient: actionClient,
	}, nil
}

func (b *Bridge) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := b.tgBotAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		logrus.Debugf("Received [%s] %s", update.Message.From.UserName, update.Message.Text)

		wResponse, err := Process(b.witClient, GetSession(&update), update.Message.Text)

		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
			//msg.ReplyToMessageID = update.Message.MessageID
			b.tgBotAPI.Send(msg)
		}

		for _, op := range wResponse.ops {
			switch op.(type) {
			case WitgoAction:
				msg, err := b.actionClient.doAction(string(op.(WitgoAction)), nil)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error: %v", err.Error()))
					//msg.ReplyToMessageID = update.Message.MessageID
					b.tgBotAPI.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
					//msg.ReplyToMessageID = update.Message.MessageID
					b.tgBotAPI.Send(msg)
				}
			case WitgoMessage:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, op.(string))
				//msg.ReplyToMessageID = update.Message.MessageID
				b.tgBotAPI.Send(msg)
			case WitgoMerge:
				logrus.Infof("Merge: %v", op)
			}
		}
	}

	return nil
}

func GetSession(update *tgbotapi.Update) *witgo.Session {
	sessionId := fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02"), strconv.FormatInt(update.Message.Chat.ID, 10))
	logrus.Debugf("SessionID: %v", sessionId)
	return witgo.NewSession(witgo.SessionID(sessionId))
}