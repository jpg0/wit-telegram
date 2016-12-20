package main

import (
	"strings"
	"github.com/kurrik/witgo/v1/witgo"
	"github.com/Sirupsen/logrus"
	"encoding/json"
)

type WitgoResponse struct {
	session *witgo.Session
	ops     []interface{}
}

type WitgoAction string
type WitgoMessage string
type WitgoMerge witgo.EntityMap

func Process(client *witgo.Client, session *witgo.Session, q string) (rv *WitgoResponse, err error) {
	var (
		response *witgo.Response
		converse *witgo.ConverseResponse
		done bool = false
	)

	rv = &WitgoResponse{ops: make([]interface{}, 1)}

	for !done {

		logrus.Debugf("Calling Wit.ai...")

		if response, err = client.Converse(session.ID(), q, session.Context); err != nil {
			logrus.Errorf("Failed to call Wit.ai: ", err)
			return
		}
		if err = response.Parse(&converse); err != nil {
			logrus.Errorf("Failed to parse response to Wit.ai: ", err)
			return
		}

		op := strings.ToLower(converse.Type)

		logrus.Debugf("Added operation: %v", op)

		switch op {
		case "action":
			rv.ops = append(rv.ops, WitgoAction(converse.Action))
		case "msg":
			rv.ops = append(rv.ops, WitgoMessage(converse.Msg))
		case "merge":
			rv.ops = append(rv.ops, WitgoMerge(converse.Entities))
		case "error":
			errString, _ := json.Marshal(converse)
			logrus.Errorf("Error from wit.ai: %v", string(errString))
			done = true
		case "stop":
			done = true
		default:
			done = true
		}
		q = ""
	}
	return
}