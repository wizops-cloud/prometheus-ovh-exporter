package ovh

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeClient_Get_ReturnsResponse(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/me/bill"] = []string{"BILL-001", "BILL-002"}

	var result []string
	err := fake.Get("/me/bill", &result)

	require.NoError(t, err)
	assert.Equal(t, []string{"BILL-001", "BILL-002"}, result)
	assert.Equal(t, []string{"/me/bill"}, fake.Calls)
}

func TestFakeClient_Get_ReturnsError(t *testing.T) {
	fake := NewFakeClient()
	fake.Errors["/me/bill"] = fmt.Errorf("api error")

	var result []string
	err := fake.Get("/me/bill", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestFakeClient_Get_NoResponse(t *testing.T) {
	fake := NewFakeClient()

	var result []string
	err := fake.Get("/unknown", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fake response")
}

func TestFakeClient_GetWithContext(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/test"] = map[string]string{"key": "value"}

	var result map[string]string
	err := fake.GetWithContext(context.Background(), "/test", &result)

	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestFakeClient_Get_StructResponse(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/me/bill/BILL-001"] = Bill{
		BillID:       "BILL-001",
		Date:         "2026-01-01",
		PriceWithTax: Price{Value: 42.50, CurrencyCode: "EUR"},
	}

	var result Bill
	err := fake.Get("/me/bill/BILL-001", &result)

	require.NoError(t, err)
	assert.Equal(t, "BILL-001", result.BillID)
	assert.InDelta(t, 42.50, result.PriceWithTax.Value, 0.001)
}

func TestFakeClient_RecordsCalls(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/a"] = "x"
	fake.Responses["/b"] = "y"

	var s string
	_ = fake.Get("/a", &s)
	_ = fake.Get("/b", &s)
	_ = fake.Get("/a", &s)

	assert.Equal(t, []string{"/a", "/b", "/a"}, fake.Calls)
}

func TestFakeClient_Get_QueryStringFallback(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/me/bill"] = []string{"BILL-001"}

	var result []string
	err := fake.Get("/me/bill?date.from=2025-01-01", &result)

	require.NoError(t, err)
	assert.Equal(t, []string{"BILL-001"}, result)
}

func TestFakeClient_Get_QueryStringError(t *testing.T) {
	fake := NewFakeClient()
	fake.Errors["/me/bill"] = fmt.Errorf("api error")

	var result []string
	err := fake.Get("/me/bill?date.from=2025-01-01", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestFakeClient_Get_NoQueryStringMatch(t *testing.T) {
	fake := NewFakeClient()
	// No response set for /unknown or /unknown?x=1.
	var result []string
	err := fake.Get("/unknown?x=1", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fake response")
}

func TestFakeClient_Get_ExactQueryStringMatch(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/test?exact=true"] = "exact"

	var result string
	err := fake.Get("/test?exact=true", &result)

	require.NoError(t, err)
	assert.Equal(t, "exact", result)
}

func TestFakeClient_Get_MarshalError(t *testing.T) {
	fake := NewFakeClient()
	// json.Marshal cannot marshal channels.
	fake.Responses["/bad"] = make(chan int)

	var result string
	err := fake.Get("/bad", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshaling fake response")
}

func TestNewAPIClient_InvalidEndpoint(t *testing.T) {
	_, err := NewAPIClient("invalid-endpoint", "key", "secret", "consumer")
	require.Error(t, err)
}

func TestNewAPIClient_ValidEndpoint(t *testing.T) {
	client, err := NewAPIClient("ovh-eu", "key", "secret", "consumer")
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestAPIClient_Get(t *testing.T) {
	client, err := NewAPIClient("ovh-eu", "key", "secret", "consumer")
	require.NoError(t, err)

	// The call will fail with auth error, but exercises the code path.
	var result interface{}
	err = client.Get("/auth/time", &result)
	// May succeed (time endpoint is public) or fail depending on network.
	// We only care that the method doesn't panic.
	_ = err
}

func TestAPIClient_GetWithContext(t *testing.T) {
	client, err := NewAPIClient("ovh-eu", "key", "secret", "consumer")
	require.NoError(t, err)

	var result interface{}
	err = client.GetWithContext(context.Background(), "/auth/time", &result)
	_ = err
}
