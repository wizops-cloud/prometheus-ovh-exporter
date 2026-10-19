package ovh

import (
	"context"
)

// Client is the interface for OVH API calls.
// Collectors depend only on this interface, never on concrete implementations.
type Client interface {
	Get(path string, result interface{}) error
	GetWithContext(ctx context.Context, path string, result interface{}) error
}
