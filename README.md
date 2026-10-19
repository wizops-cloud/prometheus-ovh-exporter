# prometheus-ovh-exporter

A Prometheus exporter for OVH Cloud metrics — billing, consumption, cloud usage, quotas, instances, storage, and savings plans.

## Quick Start

### Binary

```bash
export OVH_EXPORTER_ENDPOINT=ovh-eu
export OVH_EXPORTER_APPLICATION_KEY=your-app-key
export OVH_EXPORTER_APPLICATION_SECRET=your-app-secret
export OVH_EXPORTER_CONSUMER_KEY=your-consumer-key

./prometheus-ovh-exporter
```

### Docker

```bash
docker run -p 9100:9100 \
  -e OVH_EXPORTER_ENDPOINT=ovh-eu \
  -e OVH_EXPORTER_APPLICATION_KEY=your-app-key \
  -e OVH_EXPORTER_APPLICATION_SECRET=your-app-secret \
  -e OVH_EXPORTER_CONSUMER_KEY=your-consumer-key \
  ghcr.io/wizops-cloud/prometheus-ovh-exporter:latest
```

## Configuration

All configuration is via environment variables with the `OVH_EXPORTER_` prefix.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OVH_EXPORTER_ENDPOINT` | Yes | — | API endpoint (see [Supported endpoints](#supported-endpoints)) |
| `OVH_EXPORTER_APPLICATION_KEY` | Yes | — | OVH application key |
| `OVH_EXPORTER_APPLICATION_SECRET` | Yes | — | OVH application secret |
| `OVH_EXPORTER_CONSUMER_KEY` | Yes | — | OVH consumer key |
| `OVH_EXPORTER_CLOUD_PROJECT_IDS` | No | — | Comma-separated list of cloud project IDs |
| `OVH_EXPORTER_COLLECTORS` | No | all | Comma-separated list of collectors to enable |
| `OVH_EXPORTER_PORT` | No | `9100` | HTTP server port |
| `OVH_EXPORTER_METRICS_PATH` | No | `/metrics` | Metrics endpoint path |
| `OVH_EXPORTER_LOG_LEVEL` | No | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `OVH_EXPORTER_LOG_FORMAT` | No | `json` | Log format (`json`, `text`) |
| `OVH_EXPORTER_CACHE_TTL` | No | `15m` | Cache TTL for API responses |
| `OVH_EXPORTER_BILLING_MONTHS` | No | `12` | Number of months of billing history |

### Supported endpoints

The exporter works with all OVH group brands, including Kimsufi and So you Start. Each brand has its own API credentials.

| Endpoint | Brand | API URL |
|----------|-------|---------|
| `ovh-eu` | OVH Europe | `https://eu.api.ovh.com/1.0` |
| `ovh-ca` | OVH Canada | `https://ca.api.ovh.com/1.0` |
| `ovh-us` | OVH US | `https://api.us.ovhcloud.com/1.0` |
| `kimsufi-eu` | Kimsufi Europe | `https://eu.api.kimsufi.com/1.0` |
| `kimsufi-ca` | Kimsufi Canada | `https://ca.api.kimsufi.com/1.0` |
| `soyoustart-eu` | So you Start Europe | `https://eu.api.soyoustart.com/1.0` |
| `soyoustart-ca` | So you Start Canada | `https://ca.api.soyoustart.com/1.0` |

To monitor multiple accounts (e.g. OVH + Kimsufi), deploy one exporter instance per account with different credentials and ports.

## Collectors

| Name | Metrics | Description |
|------|---------|-------------|
| `billing` | `ovh_billing_amount_euros`, `ovh_billing_amount_without_tax_euros`, `ovh_billing_tax_euros`, `ovh_billing_bills_total` | Invoice amounts and count |
| `consumption` | `ovh_consumption_current_euros` | Current consumption by service type |
| `cloud_usage` | `ovh_cloud_usage_cost_euros` | Cloud usage cost by resource type |
| `cloud_quota` | `ovh_cloud_quota_max`, `ovh_cloud_quota_used` | Cloud quotas (cores, instances, RAM) |
| `cloud_instance` | `ovh_cloud_instances_count` | Instance count by region/flavor/status |
| `cloud_storage` | `ovh_cloud_storage_bytes`, `ovh_cloud_storage_objects` | Storage containers size and object count |
| `savings_plan` | `ovh_savings_plan_active`, `ovh_savings_plan_usage_ratio` | Savings plan status and utilization |

### Exporter Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ovh_up` | Gauge | 1 if OVH API is reachable |
| `ovh_scrape_duration_seconds` | Gauge | Scrape duration per collector |
| `ovh_scrape_errors_total` | Counter | Scrape errors per collector |

## HTTP Endpoints

| Path | Description |
|------|-------------|
| `/` | Landing page |
| `/metrics` | Prometheus metrics |
| `/-/healthy` | Health check (always 200) |
| `/-/ready` | Readiness (200 if OVH API accessible) |

## Kubernetes

### ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ovh-exporter
spec:
  selector:
    matchLabels:
      app: ovh-exporter
  endpoints:
    - port: metrics
      interval: 5m
```

## Development

```bash
make build    # Build binary
make test     # Run tests (enforces 100% coverage)
make lint     # Run golangci-lint
make docker   # Build Docker image
```

## License

Apache License 2.0
