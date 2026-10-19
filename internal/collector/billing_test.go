package collector

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

func TestBillingCollector_Name(t *testing.T) {
	c := NewBillingCollector(ovh.NewFakeClient(), 12)
	assert.Equal(t, "billing", c.Name())
}

func TestBillingCollector_Collect(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/me/bill"] = []string{"BILL-001", "BILL-002"}
	fake.Responses["/me/bill/BILL-001"] = ovh.Bill{
		BillID:          "BILL-001",
		Date:            "2026-01-15T00:00:00+01:00",
		PriceWithTax:    ovh.Price{Value: 42.50, CurrencyCode: "EUR"},
		PriceWithoutTax: ovh.Price{Value: 35.42, CurrencyCode: "EUR"},
		Tax:             ovh.Price{Value: 7.08, CurrencyCode: "EUR"},
	}
	fake.Responses["/me/bill/BILL-002"] = ovh.Bill{
		BillID:          "BILL-002",
		Date:            "2026-02-15T00:00:00+01:00",
		PriceWithTax:    ovh.Price{Value: 99.99, CurrencyCode: "EUR"},
		PriceWithoutTax: ovh.Price{Value: 83.33, CurrencyCode: "EUR"},
		Tax:             ovh.Price{Value: 16.66, CurrencyCode: "EUR"},
	}

	collector := NewBillingCollector(fake, 12)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_billing_amount_euros")
	assert.Contains(t, names, "ovh_billing_amount_without_tax_euros")
	assert.Contains(t, names, "ovh_billing_tax_euros")
	assert.Contains(t, names, "ovh_billing_bills_total")

	assertGaugeValue(t, metrics, "ovh_billing_bills_total", nil, 2)
	assertGaugeValue(t, metrics, "ovh_billing_amount_euros", map[string]string{"bill_id": "BILL-001"}, 42.50)
	assertGaugeValue(t, metrics, "ovh_billing_amount_without_tax_euros", map[string]string{"bill_id": "BILL-001"}, 35.42)
	assertGaugeValue(t, metrics, "ovh_billing_tax_euros", map[string]string{"bill_id": "BILL-001"}, 7.08)
	assertGaugeValue(t, metrics, "ovh_billing_amount_euros", map[string]string{"bill_id": "BILL-002"}, 99.99)
}

func TestBillingCollector_APIError_BillList(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Errors["/me/bill"] = fmt.Errorf("unauthorized")

	collector := NewBillingCollector(fake, 12)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	// Should still return bills_total = 0.
	assertGaugeValue(t, metrics, "ovh_billing_bills_total", nil, 0)
}

func TestBillingCollector_APIError_BillDetail(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/me/bill"] = []string{"BILL-001"}
	fake.Errors["/me/bill/BILL-001"] = fmt.Errorf("not found")

	collector := NewBillingCollector(fake, 12)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	// Bill that errored should not appear, but total should reflect attempted count.
	assertGaugeValue(t, metrics, "ovh_billing_bills_total", nil, 1)
}

func TestBillingCollector_EmptyBills(t *testing.T) {
	fake := ovh.NewFakeClient()
	fake.Responses["/me/bill"] = []string{}

	collector := NewBillingCollector(fake, 12)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)
	metrics, err := reg.Gather()
	require.NoError(t, err)

	assertGaugeValue(t, metrics, "ovh_billing_bills_total", nil, 0)
}

// assertGaugeValue finds a metric family by name and checks that a gauge with the given labels has the expected value.
func assertGaugeValue(t *testing.T, families []*dto.MetricFamily, name string, labels map[string]string, expected float64) {
	t.Helper()
	for _, fam := range families {
		if fam.GetName() != name {
			continue
		}
		for _, m := range fam.GetMetric() {
			if matchLabels(m, labels) {
				require.NotNil(t, m.GetGauge(), "metric %s is not a gauge", name)
				assert.InDelta(t, expected, m.GetGauge().GetValue(), 0.001, "metric %s value mismatch", name)
				return
			}
		}
	}
	t.Errorf("metric %s with labels %v not found", name, labels)
}

func matchLabels(m *dto.Metric, labels map[string]string) bool {
	if len(labels) == 0 {
		return len(m.GetLabel()) == 0 || labels == nil
	}
	metricLabels := make(map[string]string)
	for _, l := range m.GetLabel() {
		metricLabels[l.GetName()] = l.GetValue()
	}
	for k, v := range labels {
		if metricLabels[k] != v {
			return false
		}
	}
	return true
}
