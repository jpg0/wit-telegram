package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/jpg0/witgo/v1/witgo"
	"strings"
	"encoding/json"
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"time"
	"strconv"
	"github.com/juju/errors"
)

type Chat struct {
	b *Bridge
	chatId int64
}

func NewChat(b *Bridge, chatId int64) *Chat {
	return &Chat {
		b:b,
		chatId:chatId,
	}
}

func (c *Chat) SendMessage(text string) {
	msg := tgbotapi.NewMessage(c.chatId, text)
	c.b.tgBotAPI.Send(msg)}

func (c *Chat) DoAction(name string, entities witgo.EntityMap) error {
	ctx := c.b.GetContext(c.chatId)
	newCtx, err := c.b.actionClient.doAction(name, entities, ctx)

	if err != nil {
		return errors.Annotate(err, "Action failed")
	} else {
		logrus.Debugf("Setting context to: %+v", newCtx)
		c.b.SetContext(c.chatId, newCtx)
		return nil
	}
}

func (c *Chat) GetSessionId() string {
	sessionId := fmt.Sprintf("%v-%v-%v", c.b.sessionSeed, time.Now().Format("2006-01-02"), strconv.FormatInt(c.chatId, 10))
	logrus.Debugf("SessionID: %v", sessionId)
	return sessionId
}

func (c *Chat) GetContext() map[string]string {
	return c.b.GetContext(c.chatId)
}

func (c *Chat) Process(client *witgo.Client, q string) WitOperation {

	var converse *witgo.ConverseResponse

	logrus.Debugf("Calling Wit.ai...")

	response, err := client.Converse(witgo.SessionID(c.GetSessionId()), q, c. GetContext())

	if err != nil {
		logrus.Errorf("Failed to call Wit.ai: ", err)
		return &WitError{message:err.Error()}
	}

	if err = response.Parse(&converse); err != nil {
		logrus.Errorf("Failed to parse response to Wit.ai: ", err)
		return &WitError{message:err.Error()}
	}

	op := strings.ToLower(converse.Type)

	logrus.Debugf("Added operation: %v", op)

	switch op {
	case "action":
		return &WitAction{name: converse.Action, entityMap:converse.Entities}
	case "msg":
		return &WitMessage{text:converse.Msg}
	case "merge":
		return &WitMerge{entityMap:converse.Entities}
	case "stop":
		return &WitStop{}
	default:
		errString, err := json.Marshal(converse)

		if err != nil {
			logrus.Errorf("Error from wit.ai: %v", err)
			return &WitError{message:err.Error()}
		}

		return &WitError{message:string(errString)}
	}
}
