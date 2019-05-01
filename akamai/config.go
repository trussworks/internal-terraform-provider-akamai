package akamai

import (
	"github.com/trussworks/akamai-sdk-go/akamai"
)

// Config holds the configuration for make Akamai requests
type Config struct {
	AccessToken  string
	ClientSecret string
	ClientToken  string
	Host         string
	EdgercFile   string
}

// AkamaiClient holds our connection to Akamai.
type AkamaiClient struct {
	client *akamai.Client
}

// Client configures and returns an initialized AkamaiClient
func (c *Config) Client() (interface{}, error) {
	ac, err := akamai.NewClient(nil, nil)
	if err != nil {
		return nil, err
	}

	client := &AkamaiClient{
		client: ac,
	}

	return client, nil
}
