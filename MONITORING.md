# Monitoring, Logging, and Tracing Setup

This document describes the observability stack implementation for the microservices application.

## Components

### 1. **Jaeger** - Distributed Tracing
- **URL**: http://localhost:16686
- **Purpose**: Visualize distributed traces across all microservices
- **Implementation**: OpenTelemetry with Jaeger exporter
- All services are instrumented with OpenTelemetry for automatic tracing

### 2. **Prometheus** - Metrics Collection
- **URL**: http://localhost:9090
- **Purpose**: Collect and store time-series metrics
- **Metrics Collected**:
  - Application metrics from all services (`/metrics` endpoint)
  - Host machine metrics (CPU, RAM, filesystem, network) via Node Exporter
  - Container metrics (CPU, RAM, filesystem, network) via cAdvisor

### 3. **Loki** - Log Aggregation
- **URL**: http://localhost:3100
- **Purpose**: Aggregate and store logs from all services
- **Implementation**: Fluent-bit collects logs and forwards to Loki
- All services use structured JSON logging (logrus)

### 4. **Grafana** - Visualization
- **URL**: http://localhost:3000
- **Credentials**: admin/admin
- **Purpose**: Unified dashboard for metrics, logs, and traces
- **Datasources**:
  - Prometheus (metrics)
  - Loki (logs)
  - Jaeger (traces)

### 5. **Node Exporter** - Host Metrics
- **URL**: http://localhost:9100
- **Purpose**: Expose host machine metrics (CPU, RAM, filesystem, network)

### 6. **cAdvisor** - Container Metrics
- **URL**: http://localhost:8085
- **Purpose**: Expose Docker container metrics (CPU, RAM, filesystem, network)

### 7. **Fluent-bit** - Log Collector
- **Purpose**: Collect logs from Docker containers and forward to Loki

## Services Instrumentation

All microservices are instrumented with:

1. **OpenTelemetry Tracing**
   - Automatic HTTP request tracing
   - gRPC request tracing (for services with gRPC)
   - Custom span creation support
   - Trace context propagation

2. **Prometheus Metrics**
   - HTTP request metrics (duration, status codes, etc.)
   - gRPC request metrics
   - Custom application metrics available at `/metrics` endpoint

3. **Structured JSON Logging**
   - All logs output in JSON format
   - Contextual information (service name, action, timestamps)
   - Automatic collection by Fluent-bit

## Instrumented Services

- ✅ **blog-service** (HTTP + MongoDB)
- ✅ **stakeholders-service** (HTTP + gRPC + MongoDB)
- ✅ **follower-service** (HTTP + gRPC + Neo4j)
- ✅ **tour-service** (HTTP + MongoDB)
- ✅ **gateway-service** (HTTP + gRPC clients)

## How to Use

### Starting the Stack

```bash
# Start all services including monitoring stack
docker-compose up -d

# View logs
docker-compose logs -f

# Check service health
curl http://localhost:8080/health
```

### Viewing Traces

1. Open Jaeger UI: http://localhost:16686
2. Select a service from dropdown (e.g., "blog-service")
3. Click "Find Traces" to see all traces
4. Click on a trace to see detailed spans and timing

### Viewing Metrics

1. Open Prometheus: http://localhost:9090
2. Example queries:
   - CPU usage: `rate(container_cpu_usage_seconds_total[5m])`
   - Memory usage: `container_memory_usage_bytes`
   - HTTP requests: `http_requests_total`
   - Node CPU: `rate(node_cpu_seconds_total[5m])`

### Viewing Logs

1. Open Grafana: http://localhost:3000
2. Navigate to "Explore"
3. Select "Loki" datasource
4. Example LogQL queries:
   - All logs: `{job="fluentbit"}`
   - Service logs: `{container_name="blog-service"}`
   - Error logs: `{job="fluentbit"} |= "error"`
   - JSON field filter: `{job="fluentbit"} | json | level="error"`

### Creating Grafana Dashboards

1. Login to Grafana: http://localhost:3000 (admin/admin)
2. Create new dashboard
3. Add panels with:
   - **Prometheus queries** for metrics
   - **Loki queries** for logs
   - **Jaeger queries** for traces

## Metrics Exposed

### Host Metrics (Node Exporter)
- `node_cpu_seconds_total` - CPU usage
- `node_memory_MemTotal_bytes` - Total RAM
- `node_memory_MemAvailable_bytes` - Available RAM
- `node_filesystem_size_bytes` - Filesystem size
- `node_network_receive_bytes_total` - Network received
- `node_network_transmit_bytes_total` - Network transmitted

### Container Metrics (cAdvisor)
- `container_cpu_usage_seconds_total` - Container CPU usage
- `container_memory_usage_bytes` - Container memory usage
- `container_fs_usage_bytes` - Container filesystem usage
- `container_network_receive_bytes_total` - Container network received
- `container_network_transmit_bytes_total` - Container network transmitted

### Application Metrics
Each service exposes metrics at `/metrics` endpoint:
- HTTP request duration
- HTTP request count by status code
- Custom application metrics

## Log Format

All services log in JSON format with the following fields:

```json
{
  "service": "blog-service",
  "action": "server_start",
  "level": "info",
  "msg": "Blog service HTTP server started",
  "time": "2025-11-22T10:30:45Z"
}
```

## Troubleshooting

### Jaeger not receiving traces
- Check environment variables are set correctly
- Verify Jaeger is running: `docker ps | grep jaeger`
- Check service logs: `docker-compose logs blog-service`

### Prometheus not scraping metrics
- Check `/metrics` endpoint is accessible: `curl http://localhost:8081/metrics`
- Verify Prometheus targets: http://localhost:9090/targets
- Check prometheus.yml configuration

### Loki not receiving logs
- Verify Fluent-bit is running: `docker ps | grep fluent-bit`
- Check Fluent-bit logs: `docker-compose logs fluent-bit`
- Verify Loki is accessible: `curl http://localhost:3100/ready`

## Architecture

```
┌─────────────┐
│   Grafana   │ ← Visualization Layer
└──────┬──────┘
       │
   ┌───┴────┬──────────┐
   │        │          │
┌──▼───┐ ┌─▼────┐ ┌───▼───┐
│Prom. │ │ Loki │ │Jaeger │ ← Data Storage
└──▲───┘ └─▲────┘ └───▲───┘
   │       │          │
   │    ┌──▼──┐       │
   │    │F.bit│       │ ← Log Collection
   │    └──▲──┘       │
   │       │          │
┌──┴───────┴──────────┴───┐
│    Microservices        │ ← Application Layer
│ (blog, stakeholders,    │
│  follower, tour, GW)    │
└──────────────────────────┘
   │                   │
┌──▼────┐         ┌────▼────┐
│ Node  │         │cAdvisor │ ← System Metrics
│Export.│         │         │
└───────┘         └─────────┘
```

## Configuration Files

- `monitoring/prometheus.yml` - Prometheus scrape configuration
- `monitoring/loki-config.yml` - Loki storage and ingestion configuration
- `monitoring/fluent-bit.conf` - Fluent-bit log collection configuration
- `monitoring/grafana/provisioning/datasources/datasources.yml` - Grafana datasources
