package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/juju/errors"
	"github.com/Sirupsen/logrus"
	"time"
)

type Bridge struct {
	tgBotAPI *tgbotapi.BotAPI
	witClient *witgo.Client
	actionClient ActionClient
	contexts map[int64]map[string]string
	sessionSeed int64
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
		sessionSeed: time.Now().UnixNano(),
	}, nil
}

func (b *Bridge) Reseed() {
	b.sessionSeed = time.Now().UnixNano()
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

//func (b *Bridge) MergeEntitiesToContext(chatId int64, entitiyMap witgo.EntityMap) map[string]string {
//	ctx := b.GetContext(chatId)
//
//	for key := range entitiyMap {
//		val, err := entitiyMap.FirstEntityValue(key)
//
//		if err == nil {
//			ctx[key] = val
//		}
//	}
//
//	return ctx
//}

func (b *Bridge) UpdateContext(chatId int64, addCtx map[string]string, rmCtx []string) map[string]string {
	ctx := b.GetContext(chatId)

	for _, key := range rmCtx {
		logrus.Debugf("Removing %v from context", key)
		delete(ctx, key)
	}

	for key, value := range addCtx {
		logrus.Debugf("Adding %v as %v to context", key, value)
		ctx[key] = value
	}

	return ctx
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

		for {
			cont, opError := op.Run(opClient)

			if opError != nil {
				b.Reseed() //discard now-broken wit.ai sessions
			}

			if !cont {
				break
			}

			op = chat.Process(b.witClient, "")
		}
	}

	return nil
}