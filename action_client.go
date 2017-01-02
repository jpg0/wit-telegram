package main

import (
	"github.com/juju/errors"
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
)

type ActionClient interface {
	doAction(action string, newContext map[string]string, context map[string]string) (func (map[string]string), error)
}

type RemoteActionClient struct {
	c          *http.Client
	addressUrl string
}

type ActionRequest struct {
	Name     string `json:"name,omitempty"`
	NewContext map[string]string `json:"newcontext,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

type ActionResponse struct {
	Message string `json:"message,omitempty"`
	AddContext map[string]string `json:"addcontext,omitempty"`
	RemoveContext []string `json:"removecontext,omitempty"`
	E       error `json:"error,omitempty"`
	ReplaceContext map[string]string `json:"replacecontext,omitempty"`
}

func NewRemoteActionClient(addressUrl string) *RemoteActionClient {
	// Set up a connection to the server.

	client := &http.Client{}
	return &RemoteActionClient{
		c: client,
		addressUrl: addressUrl,
	}
}

func (ac *RemoteActionClient) doAction(action string, newCtx map[string]string, ctx map[string]string) (func(map[string]string), error) {

	if action == "reset" {
		return resetAction(ctx)
	}

	a := ActionRequest{Name: action, NewContext: newCtx, Context: ctx}
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

	return func(toUpdate map[string]string){
		if ar.ReplaceContext != nil {
			logrus.Debugf("Replacing context")

			for k := range toUpdate {
				logrus.Debugf("Removing %v", k)
				delete(toUpdate, k)
			}
			for k, v := range ar.ReplaceContext {
				logrus.Debugf("Adding %v = %v", k, v)
				toUpdate[k] = v
			}
		} else {
			logrus.Debugf("Updating context")

			for _, k := range ar.RemoveContext {
				logrus.Debugf("Removing %v", k)
				delete(toUpdate, k)
			}
			for k,v := range ar.AddContext {
				logrus.Debugf("Adding %v = %v", k, v)
				toUpdate[k] = v
			}
		}
	}, nil
}

func resetAction(ctx map[string]string) (func(map[string]string), error) {
	return func(toUpdate map[string]string) {
		for k := range ctx {
			delete(toUpdate, k)
		}
	}, nil
}