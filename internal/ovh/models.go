// Package ovh provides the OVH API client interface, implementations, and response models.
package ovh

// Price represents an OVH monetary value.
type Price struct {
	CurrencyCode string  `json:"currencyCode"`
	Value        float64 `json:"value"`
	Text         string  `json:"text"`
}

// Bill represents an OVH invoice.
type Bill struct {
	BillID          string `json:"billId"`
	Date            string `json:"date"`
	PriceWithTax    Price  `json:"priceWithTax"`
	PriceWithoutTax Price  `json:"priceWithoutTax"`
	Tax             Price  `json:"tax"`
}

// ConsumptionDetail represents consumption for a service.
type ConsumptionDetail struct {
	ServiceType string  `json:"serviceType"`
	TotalPrice  float64 `json:"totalPrice"`
}

// Consumption represents the current billing forecast.
type Consumption struct {
	ForecastedMonthlyAvailableAmount Price               `json:"forecastedMonthlyAvailableAmount"`
	Details                          []ConsumptionDetail `json:"details"`
}

// CloudUsageDetail represents usage for a cloud resource.
type CloudUsageDetail struct {
	Type     string `json:"type"`
	Quantity struct {
		Unit  string  `json:"unit"`
		Value float64 `json:"value"`
	} `json:"quantity"`
	TotalPrice float64 `json:"totalPrice"`
}

// CloudUsagePeriod represents usage during a billing period.
type CloudUsagePeriod struct {
	HourlyUsage struct {
		Instance []CloudInstanceUsage `json:"instance"`
		Storage  []CloudStorageUsage  `json:"storage"`
	} `json:"hourlyUsage"`
	MonthlyUsage struct {
		Instance []CloudInstanceUsage `json:"instance"`
		Storage  []CloudStorageUsage  `json:"storage"`
	} `json:"monthlyUsage"`
	ResourcesUsage []CloudUsageDetail `json:"resourcesUsage"`
}

// CloudInstanceUsage represents instance usage details.
type CloudInstanceUsage struct {
	Reference string `json:"reference"`
	Region    string `json:"region"`
	Quantity  struct {
		Unit  string  `json:"unit"`
		Value float64 `json:"value"`
	} `json:"quantity"`
	TotalPrice float64 `json:"totalPrice"`
}

// CloudStorageUsage represents storage usage details.
type CloudStorageUsage struct {
	Region   string `json:"region"`
	Quantity struct {
		Unit  string  `json:"unit"`
		Value float64 `json:"value"`
	} `json:"quantity"`
	TotalPrice float64 `json:"totalPrice"`
}

// CloudQuota represents cloud project quotas for a region.
type CloudQuota struct {
	Region   string       `json:"region"`
	Instance QuotaDetails `json:"instance"`
	Keypair  QuotaDetails `json:"keypair"`
	Volume   QuotaDetails `json:"volume"`
}

// QuotaDetails holds max and used values for a quota.
type QuotaDetails struct {
	MaxCores      int `json:"maxCores"`
	UsedCores     int `json:"usedCores"`
	MaxInstances  int `json:"maxInstances"`
	UsedInstances int `json:"usedInstances"`
	MaxRAM        int `json:"maxRam"`
	UsedRAM       int `json:"usedRam"`
}

// CloudInstance represents a cloud instance.
type CloudInstance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
	Status string `json:"status"`
	Flavor struct {
		Name string `json:"name"`
	} `json:"flavor"`
}

// CloudStorage represents a cloud storage container.
type CloudStorage struct {
	Name          string `json:"name"`
	Region        string `json:"region"`
	StoredBytes   int64  `json:"storedBytes"`
	StoredObjects int64  `json:"storedObjects"`
}

// SavingsPlan represents a savings plan.
type SavingsPlan struct {
	ID              string `json:"id"`
	FlavorFamily    string `json:"flavorFamily"`
	Status          string `json:"status"`
	PeriodStartDate string `json:"periodStartDate"`
	PeriodEndDate   string `json:"periodEndDate"`
}

// SavingsPlanUsage represents savings plan usage for a period.
type SavingsPlanUsage struct {
	PlanID     string  `json:"planId"`
	UsedAmount Price   `json:"usedAmount"`
	PlanAmount Price   `json:"planAmount"`
	Ratio      float64 `json:"ratio"`
}
