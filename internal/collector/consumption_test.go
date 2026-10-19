package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestConsumptionCollector_Name(t *testing.T) {
	c := NewConsumptionCollector(ovh.NewFakeClient())
	assert.Equal(t, "consumption", c.Name())
}

func TestConsumptionCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/me/consumption/usage/current"] = ovh.Consumption{
		Details: []ovh.ConsumptionDetail{
			{ServiceType: "cloud", TotalPrice: 150.25},
			{ServiceType: "domain", TotalPrice: 12.00},
		},
	}

	collector := NewConsumptionCollector(fake)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_consumption_current_euros")

	assertGaugeValue(t, metrics, "ovh_consumption_current_euros", map[string]string{"service_type": "cloud"}, 150.25)
	assertGaugeValue(t, metrics, "ovh_consumption_current_euros", map[string]string{"service_type": "domain"}, 12.00)
}

func TestConsumptionCollector_APIError(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/me/consumption/usage/current"] = fmt.Errorf("api error")

	collector := NewConsumptionCollector(fake)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	// No consumption metrics should be emitted on error.
	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_consumption_current_euros")
}

func TestConsumptionCollector_EmptyDetails(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/me/consumption/usage/current"] = ovh.Consumption{
		Details: []ovh.ConsumptionDetail{},
	}

	collector := NewConsumptionCollector(fake)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_consumption_current_euros")
}
