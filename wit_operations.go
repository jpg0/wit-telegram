package main

import (
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/Sirupsen/logrus"
)

type WitOperation interface {
	Run(oc OperationClient) bool
}

type WitAction struct {
	name string
	entityMap witgo.EntityMap
}

func (op *WitAction) Run(oc OperationClient) bool {
	err := oc.DoAction(op.name, op.entityMap)

	if err != nil {
		logrus.Errorf("Failed to run action: %v", err)
		return false
	}

	return true
}

type WitMessage struct {
	text string
}

func (op *WitMessage) Run(oc OperationClient) bool {
	oc.SendMessage(op.text)
	return true
}

type WitMerge struct {
	entityMap witgo.EntityMap
}

func (op *WitMerge) Run(oc OperationClient) bool {
	logrus.Info("MERGE!")
	return true
}

type WitStop struct {
}

func (op *WitStop) Run(oc OperationClient) bool {
	return false
}

type WitError struct {
	message string
}

func (op *WitError) Run(oc OperationClient) bool {
	logrus.Error(op.message)
	return false
}