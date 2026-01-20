# Local Observability Sandbox

A complete local observability stack demonstrating logs, metrics, traces, and alerts end-to-end using Docker Compose and industry-standard open-source tools.

## 🚀 Quick Start

### 1. Start the Stack

```bash
docker compose up --build -d
```

This will start all 8 containers:
- Grafana
- Prometheus
- Alertmanager
- Loki
- Fluent Bit
- Jaeger
- OpenTelemetry Collector
- Demo Service (Go application)

### 2. Open Grafana

Navigate to http://localhost:3000


### 3. Generate Load

```bash
docker compose --profile tools run k6
```

This runs a k6 load test that generates traffic against the demo service.

### 4. Explore the Observability Stack

Once the load test is running:

1. **View the Dashboard**: Go to **Dashboards** → **Observability Demo**
2. **Click from Metric → Trace → Log**:
3. **See Alerts Fire**: Watch for alerts when error rates exceed 5% or P95 latency exceeds 200ms

## 📊 What's Included

### Infrastructure Components

| Component | Port | Purpose |
|-----------|------|---------|
| **Grafana** | 3000 | Visualization and dashboards |
| **Prometheus** | 9090 | Metrics storage and queries |
| **Alertmanager** | 9093 | Alert management |
| **Loki** | 3100 | Log aggregation |
| **Fluent Bit** | - | Log collection and forwarding |
| **Jaeger** | 16686 | Distributed tracing UI |
| **OpenTelemetry Collector** | 4317, 4318 | Telemetry ingestion (OTLP) |
| **Demo Service** | 8080 | Sample instrumented application |

### Demo Service Endpoints

- `GET /healthz` - Health check
- `GET /work` - Simulated work with random latency and errors

The service emits:
- **Traces** via OpenTelemetry (root span + nested spans)
- **Metrics** via OpenTelemetry (request rate, error rate, latency)
- **Structured JSON logs** to stdout with trace correlation

### Observability Signals

#### Metrics
- Request rate (`http_server_request_duration_seconds_count`)
- Error rate (`http_server_request_duration_seconds_count{status="5xx"}`)
- Latency percentiles (P95)

#### Traces
- Distributed traces with parent-child span relationships
- Trace context propagation
- Viewable in Jaeger UI at http://localhost:16686

#### Logs
- Structured JSON format
- Contains `trace_id` for correlation
- Queryable in Grafana via Loki

#### Alerts
Two pre-configured alerts:
1. **High Error Rate**: Fires when error rate > 5%
2. **High P95 Latency**: Fires when P95 latency > 200ms


## 🔍 How to Use

### Viewing Metrics
1. Go to Grafana at http://localhost:3000
2. Navigate to **Dashboards** → **Observability Demo**
3. See live request rate, error rate, and latency metrics

### Following a Request Journey
1. **Start with Metrics**: Click on a high latency point in the dashboard
2. **Jump to Traces**: Click "View Traces" or navigate to Jaeger
3. **Correlate Logs**: Find the `trace_id` in the span and query Loki
4. **See the Full Picture**: Logs, metrics, and traces all connected

### Triggering Alerts
The demo service randomly returns errors and adds latency. Running the k6 load test will trigger both alerts:

```bash
docker compose --profile tools run --rm k6
```

Check alerts in:
- Grafana: **Alerting** → **Alert rules**
- Alertmanager: http://localhost:9093

### Accessing Individual UIs

- **Grafana**: http://localhost:3000 
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **Alertmanager**: http://localhost:9093

## 🧹 Cleanup

Stop and remove all containers:

```bash
docker compose down -v
```

## 📋 Acceptance Checklist

- ✅ `docker compose up --build -d` works on a clean machine
- ✅ Sample service responds to `/work`
- ✅ Metrics visible in Grafana
- ✅ Traces visible and clickable
- ✅ Logs queryable in Grafana
- ✅ Logs contain `trace_id`
- ✅ Load test visibly affects metrics and traces
- ✅ Alerts fire under load
- ✅ README with clear instructions

## 📝 Notes

This is a **local sandbox** for demonstration and learning. Out of scope:
- Kubernetes deployment
- Authentication (OIDC, SSO, RBAC)
- High availability or horizontal scaling
- Long-term retention tuning
- Production-ready dashboards
- External alerting (Slack, PagerDuty)


## 🛠️ Troubleshooting


### No data in Grafana
```bash
# Verify demo service is running
curl http://localhost:8080/healthz

# Check OpenTelemetry Collector
docker compose logs otel-collector

# Generate some traffic
docker compose --profile tools run k6
```

### Alerts not firing
```bash
# Check Prometheus targets
# Go to http://localhost:9090/targets

# Check Alertmanager
# Go to http://localhost:9093
```
