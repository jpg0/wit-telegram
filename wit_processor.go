package main

import (
	"strings"
	"github.com/kurrik/witgo/v1/witgo"
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
		if response, err = client.Converse(session.ID(), q, session.Context); err != nil {
			return
		}
		if err = response.Parse(&converse); err != nil {
			return
		}
		switch strings.ToLower(converse.Type) {
		case "action":
			rv.ops = append(rv.ops, WitgoAction(converse.Action))
		case "msg":
			rv.ops = append(rv.ops, WitgoMessage(converse.Msg))
		case "merge":
			rv.ops = append(rv.ops, WitgoMerge(converse.Entities))
		case "stop":
			done = true
		default:
			done = true
		}
		q = ""
	}
	return
}