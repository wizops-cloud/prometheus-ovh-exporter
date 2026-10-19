// Package main is the entrypoint for the prometheus-ovh-exporter.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/collector"
	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/config"
	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/server"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	setupLogging(cfg.LogLevel, cfg.LogFormat)

	slog.Info("starting prometheus-ovh-exporter", "version", version, "port", cfg.Port)

	client, err := ovh.NewAPIClient(cfg.Endpoint, cfg.ApplicationKey, cfg.ApplicationSecret, cfg.ConsumerKey)
	if err != nil {
		return fmt.Errorf("creating OVH client: %w", err)
	}

	ttl, err := time.ParseDuration(cfg.CacheTTL)
	if err != nil {
		return fmt.Errorf("parsing cache TTL: %w", err)
	}
	cachedClient := ovh.NewCachedClient(client, ttl)

	reg := buildRegistry(cachedClient, cfg)

	promRegistry := prometheus.NewRegistry()
	exporter := collector.NewExporterCollector(reg.EnabledCollectors(cfg.Collectors))
	promRegistry.MustRegister(exporter)

	readyFn := func() bool {
		var result interface{}
		return cachedClient.Get("/auth/time", &result) == nil
	}

	srv := server.New(server.Config{
		MetricsPath: cfg.MetricsPath,
		Port:        cfg.Port,
	}, promRegistry, readyFn)

	httpSrv := srv.HTTPServer()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "addr", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return httpSrv.Shutdown(shutdownCtx)
}

func buildRegistry(client ovh.Client, cfg *config.Config) *collector.Registry {
	reg := collector.NewRegistry()
	reg.Register(collector.NewBillingCollector(client, cfg.BillingMonths))
	reg.Register(collector.NewConsumptionCollector(client))
	reg.Register(collector.NewCloudUsageCollector(client, cfg.CloudProjectIDs))
	reg.Register(collector.NewCloudQuotaCollector(client, cfg.CloudProjectIDs))
	reg.Register(collector.NewCloudInstanceCollector(client, cfg.CloudProjectIDs))
	reg.Register(collector.NewCloudStorageCollector(client, cfg.CloudProjectIDs))
	reg.Register(collector.NewSavingsPlanCollector(client, cfg.CloudProjectIDs))
	return reg
}

func setupLogging(level, format string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
