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
	"net/url"
	"github.com/vincent-petithory/dataurl"
)

type Chat struct {
	b      *Bridge
	chatId int64
}

func NewChat(b *Bridge, chatId int64) *Chat {
	return &Chat{
		b:b,
		chatId:chatId,
	}
}

func (c *Chat) SendMessage(text string, responses []string) {

	var msg tgbotapi.Chattable
	u, e := url.Parse(text)

	if e != nil && ((u.IsAbs() && imageSuffix(u.Path)) || u.Scheme == "data" ) {
		msg = toPhotoUpload(c.chatId, u)
	} else {
		mc := tgbotapi.NewMessage(c.chatId, text)

		if responses != nil {
			rows := make([]tgbotapi.KeyboardButton, len(responses))

			for i, response := range responses {
				rows[i] = tgbotapi.NewKeyboardButton(response)
			}

			kb := tgbotapi.NewReplyKeyboard(rows)
			kb.OneTimeKeyboard = true

			mc.ReplyMarkup = kb
		}

		msg = mc
	}

	c.b.tgBotAPI.Send(msg)
}

func toPhotoUpload(chatId int64, url *url.URL) tgbotapi.PhotoConfig {

	if url.Scheme == "data" {

		logrus.Debugf("Detected data URL, extracting data")

		dataUrl, err := dataurl.DecodeString(url.String())

		if err == nil {
			return tgbotapi.NewPhotoUpload(
				chatId,
				&tgbotapi.FileBytes{
					Bytes: dataUrl.Data,
				})
		} else {
			logrus.Errorf("Failed extracting data from URI: %v", err)
		}

	}

	return tgbotapi.NewPhotoUpload(chatId, url)
}

func imageSuffix(casedPath string) bool {
	path := strings.ToLower(casedPath)

	return strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".gif")
}

func (c *Chat) DoAction(name string, newCtx map[string]string) error {
	ctx := c.b.GetContext(c.chatId)
	updater, err := c.b.actionClient.doAction(name, newCtx, ctx)

	if err != nil {
		return errors.Annotate(err, "Action failed")
	} else {
		//logrus.Debugf("Removing context: %+v", rmCtx)
		//logrus.Debugf("Adding context: %+v", addCtx)
		c.b.UpdateContext(c.chatId, updater)
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

	response, err := client.Converse(witgo.SessionID(c.GetSessionId()), q, c.GetContext())

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
		return &WitMessage{text:converse.Msg, responses:converse.QuickReplies}
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