package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// CloudInstanceCollector collects OVH cloud instance metrics.
type CloudInstanceCollector struct {
	client     ovh.Client
	projectIDs []string
	countDesc  *prometheus.Desc
}

// NewCloudInstanceCollector creates a new CloudInstanceCollector.
func NewCloudInstanceCollector(client ovh.Client, projectIDs []string) *CloudInstanceCollector {
	return &CloudInstanceCollector{
		client:     client,
		projectIDs: projectIDs,
		countDesc: prometheus.NewDesc(
			"ovh_cloud_instances_count",
			"Number of cloud instances by status",
			[]string{"project", "region", "flavor", "status"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *CloudInstanceCollector) Name() string {
	return "cloud_instance"
}

// Describe sends metric descriptors.
func (c *CloudInstanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

// instanceKey uniquely identifies a group of instances for counting.
type instanceKey struct {
	project string
	region  string
	flavor  string
	status  string
}

// Collect gathers cloud instance metrics.
func (c *CloudInstanceCollector) Collect(ch chan<- prometheus.Metric) {
	for _, projectID := range c.projectIDs {
		var instances []ovh.CloudInstance
		path := fmt.Sprintf("/cloud/project/%s/instance", projectID)
		if err := c.client.Get(path, &instances); err != nil {
			slog.Error("failed to get cloud instances", "project", projectID, "error", err)
			continue
		}

		counts := make(map[instanceKey]float64)
		for _, inst := range instances {
			key := instanceKey{
				project: projectID,
				region:  inst.Region,
				flavor:  inst.Flavor.Name,
				status:  inst.Status,
			}
			counts[key]++
		}

		for key, count := range counts {
			ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue,
				count, key.project, key.region, key.flavor, key.status)
		}
	}
}
