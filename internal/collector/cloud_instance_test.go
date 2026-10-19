package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestCloudInstanceCollector_Name(t *testing.T) {
	c := NewCloudInstanceCollector(ovh.NewFakeClient(), []string{"proj1"})
	assert.Equal(t, "cloud_instance", c.Name())
}

func TestCloudInstanceCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/instance"] = []ovh.CloudInstance{
		{ID: "i1", Region: "GRA11", Status: "ACTIVE", Flavor: struct {
			Name string `json:"name"`
		}{Name: "b2-7"}},
		{ID: "i2", Region: "GRA11", Status: "ACTIVE", Flavor: struct {
			Name string `json:"name"`
		}{Name: "b2-7"}},
		{ID: "i3", Region: "SBG5", Status: "SHUTOFF", Flavor: struct {
			Name string `json:"name"`
		}{Name: "s1-2"}},
	}

	collector := NewCloudInstanceCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_cloud_instances_count")

	assertGaugeValue(t, metrics, "ovh_cloud_instances_count",
		map[string]string{"project": "proj1", "region": "GRA11", "flavor": "b2-7", "status": "ACTIVE"}, 2)
	assertGaugeValue(t, metrics, "ovh_cloud_instances_count",
		map[string]string{"project": "proj1", "region": "SBG5", "flavor": "s1-2", "status": "SHUTOFF"}, 1)
}

func TestCloudInstanceCollector_NoProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	collector := NewCloudInstanceCollector(fake, nil)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_instances_count")
}

func TestCloudInstanceCollector_APIError(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/cloud/project/proj1/instance"] = fmt.Errorf("forbidden")

	collector := NewCloudInstanceCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_instances_count")
}
