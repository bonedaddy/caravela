/*
Client package provides a client that allows to interact with the a CARAVELA's daemon
sending the requests to it. It can be used as a Golang SDK for the CARAVELA's.
This is the same client used in the CLI github.com/strabox/caravela/cli provided.
*/
package client

import (
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
	"net/http"
	"time"
)

// Client holds all the necessary information to interact with CARAVELA's daemon.
type Client struct {
	httpClient *http.Client   // HTTP client to send requests into CARAVELA's REST daemon
	config     *Configuration // Configuration parameters for the CARAVELA's client
}

// NewCaravelaIP creates a new client for a CARAVELA's daemon hosted in the given IP.
func NewCaravelaIP(caravelaHostIP string) *Client {
	config := DefaultConfig(caravelaHostIP)

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout(),
		},
	}

}

func NewCaravelaTimeoutIP(caravelaHostIP string, requestTimeout time.Duration) *Client {
	return &Client{
		config: DefaultConfig(caravelaHostIP),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// SubmitContainers allows to submit a set of containers that you want to deploy in the CARAVELA's system.
// The containers configurations are given by the containersConfigs slice.
func (client *Client) SubmitContainers(containersConfigs []types.ContainerConfig) *Error {
	/*
		runContainerMessage := rest.RunContainerMsg{
			ContainersConfigs: containersConfigs,
		}
	*/

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, containersConfigs, nil)
	if err != nil {
		return newClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return newClientError(errors.New("impossible deploy the container"))
	}
}

// StopContainers stops and removes all the containers given by the containersIDs slice.
func (client *Client) StopContainers(containersIDs []string) *Error {

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, containersIDs, nil)
	if err != nil {
		return newClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return newClientError(errors.New("error stopping the containers"))
	}
}

// ListContainers returns a slice of container status that represent all the containers that the user have running
// in the CARAVELA's system.
func (client *Client) ListContainers() ([]types.ContainerStatus, *Error) {
	var containersList []types.ContainerStatus

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, &containersList)
	if err != nil {
		return nil, newClientError(err)
	}

	if httpCode == http.StatusOK {
		return containersList, nil
	} else {
		return nil, newClientError(errors.New("error checking the container"))
	}
}

// Shutdown makes the daemon cleanly shutdown and leave the system.
func (client *Client) Shutdown() *Error {
	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserExitEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, nil)
	if err != nil {
		return newClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return newClientError(errors.New("error exiting from the system"))
	}
}
