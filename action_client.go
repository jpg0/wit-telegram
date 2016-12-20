package main

import (
	"github.com/kurrik/witgo/v1/witgo"
	"github.com/juju/errors"
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"fmt"
)

type ActionClient interface {
	doAction(action string, entities witgo.EntityMap) (string, error)
}

type LoggingActionClient struct {}

func (lc *LoggingActionClient) doAction(action string, entities witgo.EntityMap) (string, error) {
	logrus.Infof("Action requested: %v", action)
	return fmt.Sprintf("[Logged] Action requested: %v", action), nil
}

func NewLoggingActionClient() *LoggingActionClient {
	return &LoggingActionClient{}
}

type RemoteActionClient struct {
	c *http.Client
	addressUrl string
}

type ActionRequest struct {
	name string
	entities witgo.EntityMap
}

type ActionResponse struct {
	message string
	e error
}

func NewRemoteActionClient(addressUrl string) *RemoteActionClient {
	// Set up a connection to the server.

	client := &http.Client{}
	return &RemoteActionClient{
		c: client,
		addressUrl: addressUrl,
	}
}

func (ac *RemoteActionClient) doAction(action string, entities witgo.EntityMap) (string, error) {

	a := ActionRequest{name: action, entities: entities}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(a)

	request, err := http.NewRequest("POST", ac.addressUrl, b)

	if (err != nil) {
		return "", errors.Annotate(err, "Failed to construct remote action request")
	}

	response, err := ac.c.Do(request)

	if (err != nil) {
		return "", errors.Annotate(err, "Failed to invoke remote action")
	}

	if (response.StatusCode != http.StatusOK) {
		return "", errors.Errorf("Failed to invoke action, response code is: %v", response.StatusCode)
	}

	ar := new(ActionResponse)

	json.NewDecoder(response.Body).Decode(ar)

	return ar.message, ar.e
}