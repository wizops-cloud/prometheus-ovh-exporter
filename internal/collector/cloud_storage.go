package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// CloudStorageCollector collects OVH cloud storage metrics.
type CloudStorageCollector struct {
	client      ovh.Client
	projectIDs  []string
	bytesDesc   *prometheus.Desc
	objectsDesc *prometheus.Desc
}

// NewCloudStorageCollector creates a new CloudStorageCollector.
func NewCloudStorageCollector(client ovh.Client, projectIDs []string) *CloudStorageCollector {
	return &CloudStorageCollector{
		client:     client,
		projectIDs: projectIDs,
		bytesDesc: prometheus.NewDesc(
			"ovh_cloud_storage_bytes",
			"Storage size in bytes",
			[]string{"project", "region", "container"}, nil,
		),
		objectsDesc: prometheus.NewDesc(
			"ovh_cloud_storage_objects",
			"Number of objects in storage",
			[]string{"project", "region", "container"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *CloudStorageCollector) Name() string {
	return "cloud_storage"
}

// Describe sends metric descriptors.
func (c *CloudStorageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.bytesDesc
	ch <- c.objectsDesc
}

// Collect gathers cloud storage metrics.
func (c *CloudStorageCollector) Collect(ch chan<- prometheus.Metric) {
	for _, projectID := range c.projectIDs {
		var containers []ovh.CloudStorage
		path := fmt.Sprintf("/cloud/project/%s/storage", projectID)
		if err := c.client.Get(path, &containers); err != nil {
			slog.Error("failed to get cloud storage", "project", projectID, "error", err)
			continue
		}

		for _, s := range containers {
			ch <- prometheus.MustNewConstMetric(c.bytesDesc, prometheus.GaugeValue,
				float64(s.StoredBytes), projectID, s.Region, s.Name)
			ch <- prometheus.MustNewConstMetric(c.objectsDesc, prometheus.GaugeValue,
				float64(s.StoredObjects), projectID, s.Region, s.Name)
		}
	}
}
