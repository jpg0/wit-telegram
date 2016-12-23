package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/juju/errors"
	"github.com/Sirupsen/logrus"
)

type Bridge struct {
	tgBotAPI *tgbotapi.BotAPI
	witClient *witgo.Client
	actionClient ActionClient
	contexts map[int64]map[string]string
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
		contexts: make(map[int64]map[string]string),
	}, nil
}

//todo: expire contexts
func (b *Bridge) GetContext(chatId int64) map[string]string {
	rv := b.contexts[chatId]

	if rv == nil {
		rv = make(map[string]string)
		b.contexts[chatId] = rv
	}

	return rv
}

func (b *Bridge) SetContext(chatId int64, ctx map[string]string)  {
	b.contexts[chatId] = ctx
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

		chat := NewChat(b, update.Message.Chat.ID)
		op := chat.Process(b.witClient, update.Message.Text)

		opClient := NewCachingOperationClient(chat)

		for op.Run(opClient) {
			op = chat.Process(b.witClient, "")
		}
	}

	return nil
}