package main

import (
	"reflect"
	"github.com/Sirupsen/logrus"
	"github.com/juju/errors"
)

type OperationClient interface {
	SendMessage(text string, responses []string)
	DoAction(name string, newCtx map[string]string) error
}

type CachingOperationClient struct {
	name string
	newCtx map[string]string
	next OperationClient
}

func (coc *CachingOperationClient) SendMessage(text string, responses []string) {
	coc.next.SendMessage(text, responses)
}
func (coc *CachingOperationClient) DoAction(name string, newCtx map[string]string) error {
	if coc.name == name && reflect.DeepEqual(coc.newCtx, newCtx) {
		logrus.Errorf("Detected repeated action call to %v. Aborting.", name)
		return errors.Errorf("Repeated action call to %v", name)
	} else {
		logrus.Debugf("Running (new) remote action %v", name)
		coc.name = name
		coc.newCtx = newCtx
		return coc.next.DoAction(name, newCtx)
	}
}

func NewCachingOperationClient(operationClient OperationClient) *CachingOperationClient {
	return &CachingOperationClient{next: operationClient}
}