package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// ConsumptionCollector collects OVH consumption metrics.
type ConsumptionCollector struct {
	client      ovh.Client
	currentDesc *prometheus.Desc
}

// NewConsumptionCollector creates a new ConsumptionCollector.
func NewConsumptionCollector(client ovh.Client) *ConsumptionCollector {
	return &ConsumptionCollector{
		client: client,
		currentDesc: prometheus.NewDesc(
			"ovh_consumption_current_euros",
			"Current consumption in euros by service type",
			[]string{"service_type"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *ConsumptionCollector) Name() string {
	return "consumption"
}

// Describe sends metric descriptors.
func (c *ConsumptionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.currentDesc
}

// Collect gathers consumption metrics from the OVH API.
func (c *ConsumptionCollector) Collect(ch chan<- prometheus.Metric) {
	var consumption ovh.Consumption
	if err := c.client.Get("/me/consumption/usage/current", &consumption); err != nil {
		slog.Error("failed to get consumption", "error", err)
		return
	}

	for _, detail := range consumption.Details {
		ch <- prometheus.MustNewConstMetric(c.currentDesc, prometheus.GaugeValue, detail.TotalPrice, detail.ServiceType)
	}
}
