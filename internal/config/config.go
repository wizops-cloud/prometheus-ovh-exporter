// Package config handles configuration loading and validation for the OVH exporter.
package config

import (
	"fmt"
	"slices"

	"github.com/kelseyhightower/envconfig"
)

// ValidCollectors lists all known collector names.
var ValidCollectors = []string{
	"billing", "consumption", "cloud_usage", "cloud_quota",
	"cloud_instance", "cloud_storage", "savings_plan",
}

var validLogLevels = []string{"debug", "info", "warn", "error"}

var validLogFormats = []string{"json", "text"}

// Config holds all configuration for the exporter.
type Config struct {
	Endpoint          string   `envconfig:"ENDPOINT"           required:"true"`
	ApplicationKey    string   `envconfig:"APPLICATION_KEY"    required:"true"`
	ApplicationSecret string   `envconfig:"APPLICATION_SECRET" required:"true"`
	ConsumerKey       string   `envconfig:"CONSUMER_KEY"       required:"true"`
	CloudProjectIDs   []string `envconfig:"CLOUD_PROJECT_IDS"`
	Collectors        []string `envconfig:"COLLECTORS"`
	Port              int      `envconfig:"PORT"               default:"9100"`
	MetricsPath       string   `envconfig:"METRICS_PATH"       default:"/metrics"`
	LogLevel          string   `envconfig:"LOG_LEVEL"          default:"info"`
	LogFormat         string   `envconfig:"LOG_FORMAT"         default:"json"`
	CacheTTL          string   `envconfig:"CACHE_TTL"          default:"15m"`
	BillingMonths     int      `envconfig:"BILLING_MONTHS"     default:"12"`
}

// Load reads configuration from environment variables with the OVH_EXPORTER_ prefix.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("OVH_EXPORTER", &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Collectors) == 0 {
		cfg.Collectors = make([]string, len(ValidCollectors))
		copy(cfg.Collectors, ValidCollectors)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Port)
	}

	if c.BillingMonths < 1 {
		return fmt.Errorf("invalid billing_months: %d (must be >= 1)", c.BillingMonths)
	}

	if !slices.Contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("invalid log_level: %q (must be one of %v)", c.LogLevel, validLogLevels)
	}

	if !slices.Contains(validLogFormats, c.LogFormat) {
		return fmt.Errorf("invalid log_format: %q (must be one of %v)", c.LogFormat, validLogFormats)
	}

	for _, col := range c.Collectors {
		if !slices.Contains(ValidCollectors, col) {
			return fmt.Errorf("invalid collector: %q (must be one of %v)", col, ValidCollectors)
		}
	}

	return nil
}

// CollectorEnabled returns true if the given collector name is in the enabled list.
func (c *Config) CollectorEnabled(name string) bool {
	return slices.Contains(c.Collectors, name)
}
