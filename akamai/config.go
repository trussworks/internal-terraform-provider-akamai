package akamai

import (
	"fmt"

	"github.com/trussworks/akamai-sdk-go/akamai"
	"github.com/trussworks/akamai-sdk-go/akamai/credentials"
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
	// Configure the credentials. Check to see if we've set it in the env.
	cc := credentials.NewEnvCredentials()
	_, err := cc.Get()
	if err != nil {
		return nil, fmt.Errorf("Could not get credentials from env: %s", err)
	}

	ac, err := akamai.NewClient(nil, cc)
	if err != nil {
		return nil, err
	}

	client := &AkamaiClient{
		client: ac,
	}

	return client, nil
}
