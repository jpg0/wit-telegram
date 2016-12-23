package main

import (
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/Sirupsen/logrus"
)

type DummyActionClient struct {}

func (lc *DummyActionClient) doAction(action string, entities witgo.EntityMap, ctx map[string]string) (map[string]string, error) {
	logrus.Infof("Action requested: %v", action)

	switch action {
	case "restart":
		ctx["restarting"] = "dummy server"
	case "list":
		ctx["period"] = "10"
		ctx["data"] = "some thing\nsomething else"
	}

	return ctx, nil
}

func NewDummyActionClient() *DummyActionClient {
	return &DummyActionClient{}
}