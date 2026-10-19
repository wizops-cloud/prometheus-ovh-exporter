package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/wizops-cloud/prometheus-ovh-exporter/internal/ovh"
)

// BillingCollector collects OVH billing metrics.
type BillingCollector struct {
	client        ovh.Client
	billingMonths int

	amountDesc      *prometheus.Desc
	amountNoTaxDesc *prometheus.Desc
	taxDesc         *prometheus.Desc
	billsTotalDesc  *prometheus.Desc
}

// NewBillingCollector creates a new BillingCollector.
func NewBillingCollector(client ovh.Client, billingMonths int) *BillingCollector {
	return &BillingCollector{
		client:        client,
		billingMonths: billingMonths,
		amountDesc: prometheus.NewDesc(
			"ovh_billing_amount_euros",
			"Total amount including tax for each bill in euros",
			[]string{"bill_id", "date"}, nil,
		),
		amountNoTaxDesc: prometheus.NewDesc(
			"ovh_billing_amount_without_tax_euros",
			"Amount without tax for each bill in euros",
			[]string{"bill_id", "date"}, nil,
		),
		taxDesc: prometheus.NewDesc(
			"ovh_billing_tax_euros",
			"Tax amount for each bill in euros",
			[]string{"bill_id", "date"}, nil,
		),
		billsTotalDesc: prometheus.NewDesc(
			"ovh_billing_bills_total",
			"Total number of bills",
			nil, nil,
		),
	}
}

// Name returns the collector name.
func (b *BillingCollector) Name() string {
	return "billing"
}

// Describe sends metric descriptors.
func (b *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- b.amountDesc
	ch <- b.amountNoTaxDesc
	ch <- b.taxDesc
	ch <- b.billsTotalDesc
}

// Collect gathers billing metrics from the OVH API.
func (b *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	from := time.Now().AddDate(0, -b.billingMonths, 0).Format("2006-01-02")
	path := fmt.Sprintf("/me/bill?date.from=%s", from)

	var billIDs []string
	if err := b.client.Get(path, &billIDs); err != nil {
		slog.Error("failed to list bills", "error", err)
		ch <- prometheus.MustNewConstMetric(b.billsTotalDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(b.billsTotalDesc, prometheus.GaugeValue, float64(len(billIDs)))

	for _, billID := range billIDs {
		var bill ovh.Bill
		if err := b.client.Get(fmt.Sprintf("/me/bill/%s", billID), &bill); err != nil {
			slog.Error("failed to get bill details", "bill_id", billID, "error", err)
			continue
		}

		date := bill.Date
		ch <- prometheus.MustNewConstMetric(b.amountDesc, prometheus.GaugeValue, bill.PriceWithTax.Value, billID, date)
		ch <- prometheus.MustNewConstMetric(b.amountNoTaxDesc, prometheus.GaugeValue, bill.PriceWithoutTax.Value, billID, date)
		ch <- prometheus.MustNewConstMetric(b.taxDesc, prometheus.GaugeValue, bill.Tax.Value, billID, date)
	}
}
