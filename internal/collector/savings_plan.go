package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// SavingsPlanCollector collects OVH savings plan metrics.
type SavingsPlanCollector struct {
	client     ovh.Client
	projectIDs []string
	activeDesc *prometheus.Desc
	usageDesc  *prometheus.Desc
}

// NewSavingsPlanCollector creates a new SavingsPlanCollector.
func NewSavingsPlanCollector(client ovh.Client, projectIDs []string) *SavingsPlanCollector {
	return &SavingsPlanCollector{
		client:     client,
		projectIDs: projectIDs,
		activeDesc: prometheus.NewDesc(
			"ovh_savings_plan_active",
			"Whether the savings plan is active (1) or not (0)",
			[]string{"plan_id", "flavor_family"}, nil,
		),
		usageDesc: prometheus.NewDesc(
			"ovh_savings_plan_usage_ratio",
			"Usage ratio of the savings plan (0-1)",
			[]string{"plan_id"}, nil,
		),
	}
}

// Name returns the collector name.
func (c *SavingsPlanCollector) Name() string {
	return "savings_plan"
}

// Describe sends metric descriptors.
func (c *SavingsPlanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeDesc
	ch <- c.usageDesc
}

// Collect gathers savings plan metrics.
func (c *SavingsPlanCollector) Collect(ch chan<- prometheus.Metric) {
	for _, projectID := range c.projectIDs {
		var plans []ovh.SavingsPlan
		path := fmt.Sprintf("/cloud/project/%s/savingsPlan", projectID)
		if err := c.client.Get(path, &plans); err != nil {
			slog.Error("failed to get savings plans", "project", projectID, "error", err)
			continue
		}

		for _, plan := range plans {
			active := 0.0
			if plan.Status == "ACTIVE" {
				active = 1.0
			}
			ch <- prometheus.MustNewConstMetric(c.activeDesc, prometheus.GaugeValue,
				active, plan.ID, plan.FlavorFamily)

			var usages []ovh.SavingsPlanUsage
			usagePath := fmt.Sprintf("/cloud/project/%s/savingsPlan/%s/usage", projectID, plan.ID)
			if err := c.client.Get(usagePath, &usages); err != nil {
				slog.Error("failed to get savings plan usage", "plan_id", plan.ID, "error", err)
				continue
			}

			if len(usages) > 0 {
				ch <- prometheus.MustNewConstMetric(c.usageDesc, prometheus.GaugeValue,
					usages[len(usages)-1].Ratio, plan.ID)
			}
		}
	}
}
