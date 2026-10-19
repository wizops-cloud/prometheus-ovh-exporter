package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestSavingsPlanCollector_Name(t *testing.T) {
	c := NewSavingsPlanCollector(ovh.NewFakeClient(), []string{"proj1"})
	assert.Equal(t, "savings_plan", c.Name())
}

func TestSavingsPlanCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/savingsPlan"] = []ovh.SavingsPlan{
		{ID: "sp1", FlavorFamily: "b3", Status: "ACTIVE"},
		{ID: "sp2", FlavorFamily: "c3", Status: "TERMINATED"},
	}
	fake.Responses["/cloud/project/proj1/savingsPlan/sp1/usage"] = []ovh.SavingsPlanUsage{
		{PlanID: "sp1", Ratio: 0.85},
	}
	fake.Responses["/cloud/project/proj1/savingsPlan/sp2/usage"] = []ovh.SavingsPlanUsage{
		{PlanID: "sp2", Ratio: 0.0},
	}

	collector := NewSavingsPlanCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_savings_plan_active")
	assert.Contains(t, names, "ovh_savings_plan_usage_ratio")

	assertGaugeValue(t, metrics, "ovh_savings_plan_active",
		map[string]string{"plan_id": "sp1", "flavor_family": "b3"}, 1)
	assertGaugeValue(t, metrics, "ovh_savings_plan_active",
		map[string]string{"plan_id": "sp2", "flavor_family": "c3"}, 0)
	assertGaugeValue(t, metrics, "ovh_savings_plan_usage_ratio",
		map[string]string{"plan_id": "sp1"}, 0.85)
	assertGaugeValue(t, metrics, "ovh_savings_plan_usage_ratio",
		map[string]string{"plan_id": "sp2"}, 0.0)
}

func TestSavingsPlanCollector_NoProjects(t *testing.T) {
	fake := ovh.NewFakeClient()
	collector := NewSavingsPlanCollector(fake, nil)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_savings_plan_active")
}

func TestSavingsPlanCollector_APIError_List(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/cloud/project/proj1/savingsPlan"] = fmt.Errorf("forbidden")

	collector := NewSavingsPlanCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.NotContains(t, names, "ovh_savings_plan_active")
}

func TestSavingsPlanCollector_APIError_Usage(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/cloud/project/proj1/savingsPlan"] = []ovh.SavingsPlan{
		{ID: "sp1", FlavorFamily: "b3", Status: "ACTIVE"},
	}
	fake.Errors["/cloud/project/proj1/savingsPlan/sp1/usage"] = fmt.Errorf("not found")

	collector := NewSavingsPlanCollector(fake, []string{"proj1"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	// Active metric should still be present even if usage fails.
	assertGaugeValue(t, metrics, "ovh_savings_plan_active",
		map[string]string{"plan_id": "sp1", "flavor_family": "b3"}, 1)
}
