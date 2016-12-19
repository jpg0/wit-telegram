package main

import (
	"github.com/kurrik/witgo/v1/witgo"
	"github.com/juju/errors"
	"net/http"
	"bytes"
	"encoding/json"
)

type ActionClient struct {
	c *http.Client
	addressUrl string
}

type ActionRequest struct {
	name string
	entities witgo.EntityMap
}

func NewActionClient(addressUrl string) *ActionClient {
	// Set up a connection to the server.

	client := &http.Client{}
	return &ActionClient{
		c: client,
		addressUrl: addressUrl,
	}
}

func (ac *ActionClient) doAction(action string, entities witgo.EntityMap) error {

	a := ActionRequest{name: action, entities: entities}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(a)

	request, err := http.NewRequest("POST", ac.addressUrl, b)

	if (err != nil) {
		return errors.Annotate(err, "Failed to construct remote action request")
	}

	response, err := ac.c.Do(request)

	if (err != nil) {
		return errors.Annotate(err, "Failed to invoke remote action")
	}

	if (response.StatusCode != http.StatusOK) {
		return errors.Errorf("Failed to invoke action, response code is: %v", response.StatusCode)
	}

	return nil
}