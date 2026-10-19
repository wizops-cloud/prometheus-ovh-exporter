// Package collector provides Prometheus collectors for OVH metrics.
package collector

import (
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NamedCollector extends prometheus.Collector with a Name method.
type NamedCollector interface {
	prometheus.Collector
	Name() string
}

// Registry holds all available collectors.
type Registry struct {
	collectors []NamedCollector
}

// NewRegistry creates a new collector registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a collector to the registry.
func (r *Registry) Register(c NamedCollector) {
	r.collectors = append(r.collectors, c)
}

// Collectors returns all registered collectors.
func (r *Registry) Collectors() []NamedCollector {
	return r.collectors
}

// EnabledCollectors returns only collectors whose names are in the enabled list.
func (r *Registry) EnabledCollectors(enabled []string) []NamedCollector {
	var result []NamedCollector
	for _, c := range r.collectors {
		if slices.Contains(enabled, c.Name()) {
			result = append(result, c)
		}
	}
	return result
}

// ExporterCollector wraps multiple NamedCollectors and adds exporter-level metrics.
type ExporterCollector struct {
	collectors            []NamedCollector
	upDesc                *prometheus.Desc
	scrapeDurationDesc    *prometheus.Desc
	scrapeErrorsTotalDesc *prometheus.Desc
}

// NewExporterCollector creates an ExporterCollector wrapping the given collectors.
func NewExporterCollector(collectors []NamedCollector) *ExporterCollector {
	return &ExporterCollector{
		collectors: collectors,
		upDesc: prometheus.NewDesc(
			"ovh_up",
			"Whether the OVH API is reachable",
			nil, nil,
		),
		scrapeDurationDesc: prometheus.NewDesc(
			"ovh_scrape_duration_seconds",
			"Duration of the scrape by collector",
			[]string{"collector"}, nil,
		),
		scrapeErrorsTotalDesc: prometheus.NewDesc(
			"ovh_scrape_errors_total",
			"Total number of scrape errors by collector",
			[]string{"collector"}, nil,
		),
	}
}

// Describe sends all descriptors to the channel.
func (e *ExporterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.upDesc
	ch <- e.scrapeDurationDesc
	ch <- e.scrapeErrorsTotalDesc
	for _, c := range e.collectors {
		c.Describe(ch)
	}
}

// Collect gathers metrics from all sub-collectors with timing and error tracking.
func (e *ExporterCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	for _, c := range e.collectors {
		wg.Add(1)
		go func(col NamedCollector) {
			defer wg.Done()
			start := time.Now()
			collectCh := make(chan prometheus.Metric)

			go func() {
				defer close(collectCh)
				col.Collect(collectCh)
			}()

			for m := range collectCh {
				ch <- m
			}

			duration := time.Since(start).Seconds()
			ch <- prometheus.MustNewConstMetric(e.scrapeDurationDesc, prometheus.GaugeValue, duration, col.Name())
			ch <- prometheus.MustNewConstMetric(e.scrapeErrorsTotalDesc, prometheus.CounterValue, 0, col.Name())

			slog.Debug("collector scrape completed", "collector", col.Name(), "duration", duration)
		}(c)
	}

	wg.Wait()
	ch <- prometheus.MustNewConstMetric(e.upDesc, prometheus.GaugeValue, 1)
}
