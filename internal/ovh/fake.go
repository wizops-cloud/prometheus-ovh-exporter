package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// FakeClient implements Client for testing purposes.
// It returns predefined responses based on the request path.
type FakeClient struct {
	mu        sync.Mutex
	Responses map[string]interface{}
	Errors    map[string]error
	Calls     []string
}

// NewFakeClient creates a FakeClient with initialized maps.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		Responses: make(map[string]interface{}),
		Errors:    make(map[string]error),
	}
}

// Get records the call and returns the fake response or error for the given path.
func (f *FakeClient) Get(path string, result interface{}) error {
	return f.GetWithContext(context.Background(), path, result)
}

// GetWithContext records the call and returns the fake response or error for the given path.
func (f *FakeClient) GetWithContext(_ context.Context, path string, result interface{}) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, path)
	f.mu.Unlock()

	if err := f.findError(path); err != nil {
		return err
	}

	resp, ok := f.findResponse(path)
	if !ok {
		return fmt.Errorf("no fake response for %s", path)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshaling fake response for %s: %w", path, err)
	}

	return json.Unmarshal(data, result)
}

// findResponse looks up a response by exact match first, then by prefix (path without query string).
func (f *FakeClient) findResponse(path string) (interface{}, bool) {
	if resp, ok := f.Responses[path]; ok {
		return resp, true
	}
	base := pathWithoutQuery(path)
	if base != path {
		if resp, ok := f.Responses[base]; ok {
			return resp, true
		}
	}
	return nil, false
}

// findError looks up an error by exact match first, then by prefix (path without query string).
func (f *FakeClient) findError(path string) error {
	if err, ok := f.Errors[path]; ok {
		return err
	}
	base := pathWithoutQuery(path)
	if base != path {
		if err, ok := f.Errors[base]; ok {
			return err
		}
	}
	return nil
}

func pathWithoutQuery(path string) string {
	if i := strings.IndexByte(path, '?'); i >= 0 {
		return path[:i]
	}
	return path
}
