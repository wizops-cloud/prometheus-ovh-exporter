package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// CloudUsageCollector collects OVH cloud usage metrics.
type CloudUsageCollector struct {
	client     ovh.Client
	projectIDs []string
	costDesc   *prometheus.Desc
}

// NewCloudUsageCollector creates a new CloudUsageCollector.
func NewCloudUsageCollector(client ovh.Client, projectIDs []string) *CloudUsageCollector {
	return &CloudUsageCollector{
		client:     client,
		projectIDs: projectIDs,
		costDesc: prometheus.NewDesc(
			"ovh_cloud_usage_cost_euros",
			"Cloud usage cost in euros by resource type",
			[]string{"project", "resource_type"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *CloudUsageCollector) Name() string {
	return "cloud_usage"
}

// Describe sends metric descriptors.
func (c *CloudUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.costDesc
}

// Collect gathers cloud usage metrics.
func (c *CloudUsageCollector) Collect(ch chan<- prometheus.Metric) {
	for _, projectID := range c.projectIDs {
		var usage ovh.CloudUsagePeriod
		path := fmt.Sprintf("/cloud/project/%s/usage/current", projectID)
		if err := c.client.Get(path, &usage); err != nil {
			slog.Error("failed to get cloud usage", "project", projectID, "error", err)
			continue
		}

		for _, detail := range usage.ResourcesUsage {
			ch <- prometheus.MustNewConstMetric(c.costDesc, prometheus.GaugeValue,
				detail.TotalPrice, projectID, detail.Type)
		}
	}
}
