// Package prometheus contains code to access prometheus api.
package prometheus

import (
	"net/url"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Client is a prometheus API client·
type Client struct {
	clients map[string]v1.API
}

// New creates a new client.
func New(urls ...string) (*Client, error) {
	client := Client{
		clients: map[string]v1.API{},
	}

	for _, u := range urls {
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, err
		}

		c, err := api.NewClient(api.Config{
			Address: u,
		})

		if err != nil {
			return nil, err
		}

		client.clients[parsed.Hostname()] = v1.NewAPI(c)
	}

	return &client, nil
}
