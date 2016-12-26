package main

import (
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/Sirupsen/logrus"
	"fmt"
)

type WitOperation interface {
	Run(oc OperationClient) (bool, error)
}

type WitAction struct {
	name string
	entityMap witgo.EntityMap
}

func (op *WitAction) Run(oc OperationClient) (bool, error) {
	err := oc.DoAction(op.name, ToMap(op.entityMap))

	if err != nil {
		logrus.Errorf("Failed to run action: %v", err)
		oc.SendMessage(fmt.Sprintf("Failed to run action: %v", err), nil)
		return false, err
	}

	return true, nil
}

type WitMessage struct {
	text string
	responses []string
}

func (op *WitMessage) Run(oc OperationClient) (bool, error) {
	logrus.Debugf("Returning message: %v", op.text)
	oc.SendMessage(op.text, op.responses)
	return true, nil
}

type WitMerge struct {
	entityMap witgo.EntityMap
}

func (op *WitMerge) Run(oc OperationClient) (bool, error) {
	logrus.Info("MERGE!")
	return true, nil
}

type WitStop struct {
}

func (op *WitStop) Run(oc OperationClient) (bool, error) {
	logrus.Debugf("Stopping op chain")
	return false, nil
}

type WitError struct {
	message string
}

func (op *WitError) Run(oc OperationClient) (bool, error) {
	logrus.Errorf("WitError: %v", op.message)
	return false, nil
}

func ToMap(em witgo.EntityMap) map[string]string {
	ctx := make(map[string]string)

	for key := range em {
		val, err := em.FirstEntityValue(key)

		if err == nil {
			ctx[key] = val
		}
	}

	return ctx
}