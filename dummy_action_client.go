package main

import (
	"github.com/Sirupsen/logrus"
)

type DummyActionClient struct {}

func (lc *DummyActionClient) doAction(action string, newCtx map[string]string, ctx map[string]string) (map[string]string, []string, error) {
	logrus.Infof("Action requested: %v", action)

	addCtx := make(map[string]string)

	switch action {
	case "restart":
		addCtx["restarting"] = "dummy server"
	case "list":
		addCtx["period"] = "10"
		addCtx["data"] = "some thing\nsomething else"
	}

	return addCtx, []string{}, nil
}

func NewDummyActionClient() *DummyActionClient {
	return &DummyActionClient{}
}