package ovh

import (
	"context"
	"fmt"

	goovh "github.com/ovh/go-ovh/ovh"
)

// APIClient wraps the official go-ovh client to implement the Client interface.
type APIClient struct {
	client *goovh.Client
}

// NewAPIClient creates a new OVH API client.
func NewAPIClient(endpoint, appKey, appSecret, consumerKey string) (*APIClient, error) {
	c, err := goovh.NewClient(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		return nil, fmt.Errorf("creating OVH client: %w", err)
	}
	return &APIClient{client: c}, nil
}

// Get performs a GET request to the OVH API.
func (a *APIClient) Get(path string, result interface{}) error {
	return a.client.Get(path, result)
}

// GetWithContext performs a GET request with context.
func (a *APIClient) GetWithContext(ctx context.Context, path string, result interface{}) error {
	return a.client.GetWithContext(ctx, path, result)
}
