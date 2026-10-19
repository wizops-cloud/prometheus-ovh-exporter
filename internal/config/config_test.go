package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OVH_EXPORTER_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_EXPORTER_APPLICATION_KEY", "test-app-key")
	t.Setenv("OVH_EXPORTER_APPLICATION_SECRET", "test-app-secret")
	t.Setenv("OVH_EXPORTER_CONSUMER_KEY", "test-consumer-key")
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "ovh-eu", cfg.Endpoint)
	assert.Equal(t, "test-app-key", cfg.ApplicationKey)
	assert.Equal(t, "test-app-secret", cfg.ApplicationSecret)
	assert.Equal(t, "test-consumer-key", cfg.ConsumerKey)
	assert.Equal(t, 9100, cfg.Port)
	assert.Equal(t, "/metrics", cfg.MetricsPath)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "15m", cfg.CacheTTL)
	assert.Equal(t, 12, cfg.BillingMonths)
	assert.Empty(t, cfg.CloudProjectIDs)
	assert.Equal(t, []string{
		"billing", "consumption", "cloud_usage", "cloud_quota",
		"cloud_instance", "cloud_storage", "savings_plan",
	}, cfg.Collectors)
}

func TestLoad_CustomValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_PORT", "8080")
	t.Setenv("OVH_EXPORTER_METRICS_PATH", "/custom-metrics")
	t.Setenv("OVH_EXPORTER_LOG_LEVEL", "debug")
	t.Setenv("OVH_EXPORTER_LOG_FORMAT", "text")
	t.Setenv("OVH_EXPORTER_CACHE_TTL", "30m")
	t.Setenv("OVH_EXPORTER_BILLING_MONTHS", "6")
	t.Setenv("OVH_EXPORTER_CLOUD_PROJECT_IDS", "proj1,proj2")
	t.Setenv("OVH_EXPORTER_COLLECTORS", "billing,consumption")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "/custom-metrics", cfg.MetricsPath)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.Equal(t, "30m", cfg.CacheTTL)
	assert.Equal(t, 6, cfg.BillingMonths)
	assert.Equal(t, []string{"proj1", "proj2"}, cfg.CloudProjectIDs)
	assert.Equal(t, []string{"billing", "consumption"}, cfg.Collectors)
}

func TestLoad_MissingEndpoint(t *testing.T) {
	t.Setenv("OVH_EXPORTER_APPLICATION_KEY", "key")
	t.Setenv("OVH_EXPORTER_APPLICATION_SECRET", "secret")
	t.Setenv("OVH_EXPORTER_CONSUMER_KEY", "consumer")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ENDPOINT")
}

func TestLoad_MissingApplicationKey(t *testing.T) {
	t.Setenv("OVH_EXPORTER_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_EXPORTER_APPLICATION_SECRET", "secret")
	t.Setenv("OVH_EXPORTER_CONSUMER_KEY", "consumer")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APPLICATION_KEY")
}

func TestLoad_MissingApplicationSecret(t *testing.T) {
	t.Setenv("OVH_EXPORTER_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_EXPORTER_APPLICATION_KEY", "key")
	t.Setenv("OVH_EXPORTER_CONSUMER_KEY", "consumer")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APPLICATION_SECRET")
}

func TestLoad_MissingConsumerKey(t *testing.T) {
	t.Setenv("OVH_EXPORTER_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_EXPORTER_APPLICATION_KEY", "key")
	t.Setenv("OVH_EXPORTER_APPLICATION_SECRET", "secret")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CONSUMER_KEY")
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_LOG_LEVEL", "invalid")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log_level")
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_LOG_FORMAT", "yaml")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log_format")
}

func TestLoad_InvalidCollector(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_COLLECTORS", "billing,nonexistent")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestLoad_InvalidPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_PORT", "99999")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestLoad_InvalidBillingMonths(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OVH_EXPORTER_BILLING_MONTHS", "0")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "billing_months")
}

func TestConfig_CollectorEnabled(t *testing.T) {
	cfg := &Config{
		Collectors: []string{"billing", "consumption"},
	}

	assert.True(t, cfg.CollectorEnabled("billing"))
	assert.True(t, cfg.CollectorEnabled("consumption"))
	assert.False(t, cfg.CollectorEnabled("cloud_usage"))
}
