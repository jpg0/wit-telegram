package main

import (
	//"github.com/jpg0/witgo/v1/witgo"
	"github.com/juju/errors"
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
)

type ActionClient interface {
	doAction(action string, context map[string]string) (map[string]string, error)
}

type RemoteActionClient struct {
	c          *http.Client
	addressUrl string
}

type ActionRequest struct {
	Name     string `json:"name,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

type ActionResponse struct {
	Message string `json:"message,omitempty"`
	Context map[string]string `json:"context,omitempty"`
	E       error `json:"error,omitempty"`
}

func NewRemoteActionClient(addressUrl string) *RemoteActionClient {
	// Set up a connection to the server.

	client := &http.Client{}
	return &RemoteActionClient{
		c: client,
		addressUrl: addressUrl,
	}
}

func (ac *RemoteActionClient) doAction(action string, ctx map[string]string) (map[string]string, error) {

	a := ActionRequest{Name: action, Context: ctx}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(a)

	request, err := http.NewRequest("POST", ac.addressUrl, b)

	if (err != nil) {
		return nil, errors.Annotate(err, "Failed to construct remote action request")
	}

	response, err := ac.c.Do(request)

	if (err != nil) {
		return nil, errors.Annotate(err, "Failed to invoke remote action")
	}

	if (response.StatusCode != http.StatusOK) {

		ar := new(ActionResponse)
		err = json.NewDecoder(response.Body).Decode(ar)

		if err == nil {
			logrus.Errorf("Received remote error: %v", ar.E)
		}

		return nil, errors.Errorf("Failed to invoke action, response code is: %v", response.StatusCode)
	}

	ar := new(ActionResponse)
	json.NewDecoder(response.Body).Decode(ar)

	return ar.Context, ar.E
}