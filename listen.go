package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/kurrik/witgo/v1/witgo"
	"github.com/juju/errors"
	"github.com/Sirupsen/logrus"
	"strconv"
)

type Bridge struct {
	tgBotAPI *tgbotapi.BotAPI
	witClient *witgo.Client
}


func NewBridge(tgKey string, witKey string) (*Bridge, error) {
	bot, err := tgbotapi.NewBotAPI(tgKey)
	if err != nil {
		return nil, errors.Annotate(err, "Failed to connect to Telegram")
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	witClient := witgo.NewClient(witKey)


	return &Bridge{
		tgBotAPI: bot,
		witClient: witClient,
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

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		wResponse, err := Process(b.witClient, witgo.NewSession(witgo.SessionID(strconv.FormatInt(update.Message.Chat.ID, 10))), update.Message.Text)

		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
			msg.ReplyToMessageID = update.Message.MessageID
			b.tgBotAPI.Send(msg)
		}

		for _, op := range wResponse.ops {
			switch op.(type) {
			case WitgoAction:
				logrus.Infof("Action: %v", op)


			case WitgoMessage:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, op.(string))
				msg.ReplyToMessageID = update.Message.MessageID
				b.tgBotAPI.Send(msg)
			case WitgoMerge:
				logrus.Infof("Merge: %v", op)
			}
		}
	}

	return nil
}