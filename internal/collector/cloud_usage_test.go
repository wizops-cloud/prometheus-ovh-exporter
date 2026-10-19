package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestCloudUsageCollector_Name(t *testing.T) {
	c := NewCloudUsageCollector(ovh.NewFakeClient(), []string{"proj1"})
	assert.Equal(t, "cloud_usage", c.Name())
}

func TestCloudUsageCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/usage/current"] = ovh.CloudUsagePeriod{
		ResourcesUsage: []ovh.CloudUsageDetail{
			{Type: "instance", TotalPrice: 25.50},
			{Type: "storage", TotalPrice: 5.00},
		},
	}

	collector := NewCloudUsageCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_cloud_usage_cost_euros")

	assertGaugeValue(t, metrics, "ovh_cloud_usage_cost_euros",
		map[string]string{"project": "proj1", "resource_type": "instance"}, 25.50)
	assertGaugeValue(t, metrics, "ovh_cloud_usage_cost_euros",
		map[string]string{"project": "proj1", "resource_type": "storage"}, 5.00)
}

func TestCloudUsageCollector_MultipleProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/p1/usage/current"] = ovh.CloudUsagePeriod{
		ResourcesUsage: []ovh.CloudUsageDetail{
			{Type: "instance", TotalPrice: 10.00},
		},
	}
	fake.Responses["/cloud/project/p2/usage/current"] = ovh.CloudUsagePeriod{
		ResourcesUsage: []ovh.CloudUsageDetail{
			{Type: "storage", TotalPrice: 20.00},
		},
	}

	collector := NewCloudUsageCollector(fake, []string{"p1", "p2"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	assertGaugeValue(t, metrics, "ovh_cloud_usage_cost_euros",
		map[string]string{"project": "p1", "resource_type": "instance"}, 10.00)
	assertGaugeValue(t, metrics, "ovh_cloud_usage_cost_euros",
		map[string]string{"project": "p2", "resource_type": "storage"}, 20.00)
}

func TestCloudUsageCollector_NoProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	collector := NewCloudUsageCollector(fake, nil)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_usage_cost_euros")
}

func TestCloudUsageCollector_APIError(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/cloud/project/proj1/usage/current"] = fmt.Errorf("forbidden")

	collector := NewCloudUsageCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_usage_cost_euros")
}
