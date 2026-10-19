package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestCloudQuotaCollector_Name(t *testing.T) {
	c := NewCloudQuotaCollector(ovh.NewFakeClient(), []string{"proj1"})
	assert.Equal(t, "cloud_quota", c.Name())
}

func TestCloudQuotaCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/quota"] = []ovh.CloudQuota{
		{
			Region: "GRA11",
			Instance: ovh.QuotaDetails{
				MaxCores:      20,
				UsedCores:     4,
				MaxInstances:  10,
				UsedInstances: 2,
				MaxRAM:        40960,
				UsedRAM:       8192,
			},
		},
	}

	collector := NewCloudQuotaCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_cloud_quota_max")
	assert.Contains(t, names, "ovh_cloud_quota_used")

	assertGaugeValue(t, metrics, "ovh_cloud_quota_max",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "cores"}, 20)
	assertGaugeValue(t, metrics, "ovh_cloud_quota_used",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "cores"}, 4)
	assertGaugeValue(t, metrics, "ovh_cloud_quota_max",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "instances"}, 10)
	assertGaugeValue(t, metrics, "ovh_cloud_quota_used",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "instances"}, 2)
	assertGaugeValue(t, metrics, "ovh_cloud_quota_max",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "ram"}, 40960)
	assertGaugeValue(t, metrics, "ovh_cloud_quota_used",
		map[string]string{"project": "proj1", "region": "GRA11", "resource": "ram"}, 8192)
}

func TestCloudQuotaCollector_NoProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	collector := NewCloudQuotaCollector(fake, nil)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_quota_max")
}

func TestCloudQuotaCollector_APIError(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/cloud/project/proj1/quota"] = fmt.Errorf("forbidden")

	collector := NewCloudQuotaCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_quota_max")
}
