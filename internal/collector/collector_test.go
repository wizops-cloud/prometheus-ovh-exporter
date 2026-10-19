package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCollector struct {
	name string
}

func (s *stubCollector) Name() string { return s.name }
func (s *stubCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("stub_"+s.name, "stub", nil, nil)
}
func (s *stubCollector) Collect(ch chan<- prometheus.Metric) {}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()
	c := &stubCollector{name: "billing"}

	reg.Register(c)

	assert.Len(t, reg.Collectors(), 1)
	assert.Equal(t, "billing", reg.Collectors()[0].Name())
}

func TestRegistry_RegisterMultiple(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubCollector{name: "billing"})
	reg.Register(&stubCollector{name: "consumption"})

	assert.Len(t, reg.Collectors(), 2)
}

func TestRegistry_EnabledOnly(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubCollector{name: "billing"})
	reg.Register(&stubCollector{name: "consumption"})
	reg.Register(&stubCollector{name: "cloud_usage"})

	enabled := reg.EnabledCollectors([]string{"billing", "cloud_usage"})

	assert.Len(t, enabled, 2)

	names := make([]string, len(enabled))
	for i, c := range enabled {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "billing")
	assert.Contains(t, names, "cloud_usage")
	assert.NotContains(t, names, "consumption")
}

func TestRegistry_EnabledOnly_Empty(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubCollector{name: "billing"})

	enabled := reg.EnabledCollectors([]string{})
	assert.Empty(t, enabled)
}

// emittingCollector is a stub that actually emits a metric.
type emittingCollector struct {
	name string
	desc *prometheus.Desc
}

func newEmittingCollector(name string) *emittingCollector {
	return &emittingCollector{
		name: name,
		desc: prometheus.NewDesc("test_"+name+"_metric", "test metric", nil, nil),
	}
}

func (e *emittingCollector) Name() string { return e.name }
func (e *emittingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.desc
}
func (e *emittingCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(e.desc, prometheus.GaugeValue, 42)
}

func TestExporterCollector_DescribeAndCollect(t *testing.T) {
	c1 := &stubCollector{name: "billing"}
	c2 := &stubCollector{name: "consumption"}

	exporter := NewExporterCollector([]NamedCollector{c1, c2})

	// Test Describe.
	descCh := make(chan *prometheus.Desc, 20)
	exporter.Describe(descCh)
	close(descCh)

	descs := make([]*prometheus.Desc, 0, len(descCh))
	for d := range descCh {
		descs = append(descs, d)
	}
	// Should have sub-collector descs + ovh_up + ovh_scrape_duration_seconds + ovh_scrape_errors_total.
	require.GreaterOrEqual(t, len(descs), 4)

	// Test Collect.
	promReg := prometheus.NewPedanticRegistry()
	promReg.MustRegister(exporter)
	metrics, err := promReg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_up")
	assert.Contains(t, names, "ovh_scrape_duration_seconds")
}

func TestExporterCollector_CollectWithEmittingCollector(t *testing.T) {
	c := newEmittingCollector("test")

	exporter := NewExporterCollector([]NamedCollector{c})

	promReg := prometheus.NewPedanticRegistry()
	promReg.MustRegister(exporter)
	metrics, err := promReg.Gather()
	require.NoError(t, err)

	names := metricFamilyNames(metrics)
	assert.Contains(t, names, "ovh_up")
	assert.Contains(t, names, "test_test_metric")
	assert.Contains(t, names, "ovh_scrape_duration_seconds")
	assert.Contains(t, names, "ovh_scrape_errors_total")

	assertGaugeValue(t, metrics, "test_test_metric", nil, 42)
}

func metricFamilyNames(families []*dto.MetricFamily) []string {
	names := make([]string, len(families))
	for i, f := range families {
		names[i] = f.GetName()
	}
	return names
}
