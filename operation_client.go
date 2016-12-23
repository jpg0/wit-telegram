package main

import (
	"reflect"
	"github.com/Sirupsen/logrus"
	"github.com/jpg0/witgo/v1/witgo"
	"github.com/juju/errors"
)

type OperationClient interface {
	SendMessage(text string)
	DoAction(name string, entities witgo.EntityMap) error
}

type CachingOperationClient struct {
	name string
	entities witgo.EntityMap
	next OperationClient
}

func (coc *CachingOperationClient) SendMessage(text string) {
	coc.next.SendMessage(text)
}
func (coc *CachingOperationClient) DoAction(name string, entities witgo.EntityMap) error {
	if coc.name == name && reflect.DeepEqual(coc.entities, entities) {
		logrus.Errorf("Detected repeated action call to %v. Aborting.", name)
		return errors.Errorf("Repeated action call to %v", name)
	} else {
		logrus.Debugf("Running (new) remote action %v", name)
		coc.name = name
		coc.entities = entities
		return coc.next.DoAction(name, entities)
	}
}

func NewCachingOperationClient(operationClient OperationClient) *CachingOperationClient {
	return &CachingOperationClient{next: operationClient}
}