package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestCloudStorageCollector_Name(t *testing.T) {
	c := NewCloudStorageCollector(ovh.NewFakeClient(), []string{"proj1"})
	assert.Equal(t, "cloud_storage", c.Name())
}

func TestCloudStorageCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/storage"] = []ovh.CloudStorage{
		{Name: "backups", Region: "GRA", StoredBytes: 1073741824, StoredObjects: 42},
		{Name: "logs", Region: "SBG", StoredBytes: 536870912, StoredObjects: 100},
	}

	collector := NewCloudStorageCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_cloud_storage_bytes")
	assert.Contains(t, names, "ovh_cloud_storage_objects")

	assertGaugeValue(t, metrics, "ovh_cloud_storage_bytes",
		map[string]string{"project": "proj1", "region": "GRA", "container": "backups"}, 1073741824)
	assertGaugeValue(t, metrics, "ovh_cloud_storage_objects",
		map[string]string{"project": "proj1", "region": "GRA", "container": "backups"}, 42)
	assertGaugeValue(t, metrics, "ovh_cloud_storage_bytes",
		map[string]string{"project": "proj1", "region": "SBG", "container": "logs"}, 536870912)
	assertGaugeValue(t, metrics, "ovh_cloud_storage_objects",
		map[string]string{"project": "proj1", "region": "SBG", "container": "logs"}, 100)
}

func TestCloudStorageCollector_NoProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	collector := NewCloudStorageCollector(fake, nil)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_storage_bytes")
}

func TestCloudStorageCollector_APIError(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/cloud/project/proj1/storage"] = fmt.Errorf("forbidden")

	collector := NewCloudStorageCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_cloud_storage_bytes")
}
