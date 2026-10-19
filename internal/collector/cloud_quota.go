package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// CloudQuotaCollector collects OVH cloud quota metrics.
type CloudQuotaCollector struct {
	client     ovh.Client
	projectIDs []string
	maxDesc    *prometheus.Desc
	usedDesc   *prometheus.Desc
}

// NewCloudQuotaCollector creates a new CloudQuotaCollector.
func NewCloudQuotaCollector(client ovh.Client, projectIDs []string) *CloudQuotaCollector {
	return &CloudQuotaCollector{
		client:     client,
		projectIDs: projectIDs,
		maxDesc: prometheus.NewDesc(
			"ovh_cloud_quota_max",
			"Maximum cloud quota value",
			[]string{"project", "region", "resource"}, nil,
		),
		usedDesc: prometheus.NewDesc(
			"ovh_cloud_quota_used",
			"Used cloud quota value",
			[]string{"project", "region", "resource"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *CloudQuotaCollector) Name() string {
	return "cloud_quota"
}

// Describe sends metric descriptors.
func (c *CloudQuotaCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxDesc
	ch <- c.usedDesc
}

// Collect gathers cloud quota metrics.
func (c *CloudQuotaCollector) Collect(ch chan<- prometheus.Metric) {
	for _, projectID := range c.projectIDs {
		var quotas []ovh.CloudQuota
		path := fmt.Sprintf("/cloud/project/%s/quota", projectID)
		if err := c.client.Get(path, &quotas); err != nil {
			slog.Error("failed to get cloud quotas", "project", projectID, "error", err)
			continue
		}

		for i := range quotas {
			q := &quotas[i]
			ch <- prometheus.MustNewConstMetric(c.maxDesc, prometheus.GaugeValue,
				float64(q.Instance.MaxCores), projectID, q.Region, "cores")
			ch <- prometheus.MustNewConstMetric(c.usedDesc, prometheus.GaugeValue,
				float64(q.Instance.UsedCores), projectID, q.Region, "cores")
			ch <- prometheus.MustNewConstMetric(c.maxDesc, prometheus.GaugeValue,
				float64(q.Instance.MaxInstances), projectID, q.Region, "instances")
			ch <- prometheus.MustNewConstMetric(c.usedDesc, prometheus.GaugeValue,
				float64(q.Instance.UsedInstances), projectID, q.Region, "instances")
			ch <- prometheus.MustNewConstMetric(c.maxDesc, prometheus.GaugeValue,
				float64(q.Instance.MaxRAM), projectID, q.Region, "ram")
			ch <- prometheus.MustNewConstMetric(c.usedDesc, prometheus.GaugeValue,
				float64(q.Instance.UsedRAM), projectID, q.Region, "ram")
		}
	}
}
