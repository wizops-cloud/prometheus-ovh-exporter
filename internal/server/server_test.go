package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLandingPage(t *testing.T) {
	srv := New(Config{
		MetricsPath: "/metrics",
	}, prometheus.NewRegistry(), func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "OVH Exporter")
	assert.Contains(t, body, "/metrics")
}

func TestHealthEndpoint(t *testing.T) {
	srv := New(Config{
		MetricsPath: "/metrics",
	}, prometheus.NewRegistry(), func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/-/healthy", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Equal(t, "OK", string(body))
}

func TestReadyEndpoint_Ready(t *testing.T) {
	srv := New(Config{
		MetricsPath: "/metrics",
	}, prometheus.NewRegistry(), func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/-/ready", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Equal(t, "OK", string(body))
}

func TestReadyEndpoint_NotReady(t *testing.T) {
	srv := New(Config{
		MetricsPath: "/metrics",
	}, prometheus.NewRegistry(), func() bool { return false })

	req := httptest.NewRequest(http.MethodGet, "/-/ready", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestMetricsEndpoint(t *testing.T) {
	reg := prometheus.NewRegistry()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_metric",
		Help: "A test metric",
	})
	reg.MustRegister(gauge)
	gauge.Set(42)

	srv := New(Config{
		MetricsPath: "/metrics",
	}, reg, func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "test_metric 42")
}

func TestCustomMetricsPath(t *testing.T) {
	reg := prometheus.NewRegistry()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "custom_path_metric",
		Help: "Test",
	})
	reg.MustRegister(gauge)
	gauge.Set(1)

	srv := New(Config{
		MetricsPath: "/custom-metrics",
	}, reg, func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/custom-metrics", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "custom_path_metric 1")
}

func TestNewHTTPServer(t *testing.T) {
	reg := prometheus.NewRegistry()
	srv := New(Config{
		MetricsPath: "/metrics",
		Port:        9100,
	}, reg, func() bool { return true })

	httpSrv := srv.HTTPServer()
	require.NotNil(t, httpSrv)
	assert.Equal(t, ":9100", httpSrv.Addr)
}
