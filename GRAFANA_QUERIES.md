# Grafana Query Examples

This document provides example queries for Grafana dashboards to visualize metrics, logs, and traces.

## Prometheus Queries (Metrics)

### Host Machine Metrics

**CPU Usage by Core**
```promql
rate(node_cpu_seconds_total{mode!="idle"}[5m]) * 100
```

**Total RAM Usage**
```promql
(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100
```

**Available Memory**
```promql
node_memory_MemAvailable_bytes / 1024 / 1024 / 1024
```

**Filesystem Usage**
```promql
(node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100
```

**Network Traffic Received**
```promql
rate(node_network_receive_bytes_total[5m])
```

**Network Traffic Transmitted**
```promql
rate(node_network_transmit_bytes_total[5m])
```

### Container Metrics

**Container CPU Usage by Service**
```promql
rate(container_cpu_usage_seconds_total{name=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"}[5m]) * 100
```

**Container Memory Usage by Service**
```promql
container_memory_usage_bytes{name=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"} / 1024 / 1024
```

**Container Network Received**
```promql
rate(container_network_receive_bytes_total{name=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"}[5m])
```

**Container Network Transmitted**
```promql
rate(container_network_transmit_bytes_total{name=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"}[5m])
```

**Container Filesystem Usage**
```promql
container_fs_usage_bytes{name=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"} / 1024 / 1024 / 1024
```

### Application Metrics

**HTTP Request Rate by Service**
```promql
rate(http_requests_total[5m])
```

**HTTP Request Duration (P95)**
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**HTTP Request Duration (P99)**
```promql
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

**HTTP Errors (5xx)**
```promql
rate(http_requests_total{code=~"5.."}[5m])
```

**Active Go Goroutines**
```promql
go_goroutines
```

**Heap Memory Usage**
```promql
go_memstats_heap_alloc_bytes / 1024 / 1024
```

## Loki Queries (Logs)

### Basic Log Queries

**All logs**
```logql
{job="fluentbit"}
```

**Logs from specific service**
```logql
{container_name="blog-service"}
```

**Logs from multiple services**
```logql
{container_name=~"blog-service|gateway"}
```

### Filtered Log Queries

**Error logs only**
```logql
{job="fluentbit"} | json | level="error"
```

**Warning and error logs**
```logql
{job="fluentbit"} | json | level=~"error|warn"
```

**Logs containing specific text**
```logql
{container_name="blog-service"} |= "MongoDB"
```

**Logs NOT containing specific text**
```logql
{container_name="blog-service"} != "health"
```

**Logs with specific action**
```logql
{job="fluentbit"} | json | action="db_connect"
```

### Advanced Log Queries

**Log count by service**
```logql
sum by (container_name) (count_over_time({job="fluentbit"}[5m]))
```

**Error rate per service**
```logql
sum by (container_name) (rate({job="fluentbit"} | json | level="error" [5m]))
```

**Logs with JSON parsing**
```logql
{container_name="blog-service"} | json | service="blog-service" | action="server_start"
```

**Top 10 most common log messages**
```logql
topk(10, sum by (msg) (count_over_time({job="fluentbit"} | json [1h])))
```

## Jaeger Queries (Traces)

Jaeger queries are done through the UI:

1. **Service**: Select from dropdown (blog-service, stakeholders-service, etc.)
2. **Operation**: Select specific operation or leave blank for all
3. **Tags**: Add filters like:
   - `http.status_code=200`
   - `http.method=POST`
   - `error=true`
4. **Lookback**: Choose time range (Last Hour, Last 2 Hours, etc.)
5. **Min/Max Duration**: Filter by request duration

## Dashboard Examples

### System Overview Dashboard

Create a dashboard with:
1. **CPU Usage** - Gauge showing node_cpu percentage
2. **Memory Usage** - Gauge showing memory percentage
3. **Network Traffic** - Graph showing receive/transmit rates
4. **Disk Usage** - Gauge showing filesystem percentage
5. **Container Count** - Stat showing number of running containers

### Service Health Dashboard

Create a dashboard with:
1. **Request Rate** - Graph of HTTP requests per second by service
2. **Error Rate** - Graph of 5xx errors per second
3. **Response Time P95** - Graph of 95th percentile response times
4. **Response Time P99** - Graph of 99th percentile response times
5. **Active Connections** - Stat showing current connections

### Logs Dashboard

Create a dashboard with:
1. **Recent Errors** - Table of recent error logs
2. **Log Volume** - Graph of logs per second by level
3. **Service Activity** - Heatmap of logs by service and time
4. **Top Error Messages** - Table of most common errors

## Creating Grafana Dashboard

1. Login to Grafana (http://localhost:3000)
2. Click "+" → "Dashboard"
3. Click "Add visualization"
4. Select datasource (Prometheus, Loki, or Jaeger)
5. Enter query from examples above
6. Configure visualization type (Graph, Gauge, Table, etc.)
7. Set panel title and description
8. Click "Apply"
9. Click "Save dashboard"

## Alerting Rules (Prometheus)

Add to `prometheus.yml` under `rule_files`:

```yaml
groups:
  - name: service_alerts
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: rate(node_cpu_seconds_total{mode!="idle"}[5m]) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          
      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
          
      - alert: ServiceDown
        expr: up{job=~"blog-service|stakeholders-service|follower-service|tour-service|gateway"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
```

## Tips

1. **Use variables** in dashboards for dynamic service selection
2. **Set refresh rates** appropriate for your use case (5s, 10s, 30s)
3. **Use template variables** for filtering by environment, service, etc.
4. **Create separate dashboards** for different concerns (system, application, business)
5. **Use annotations** to mark deployments and incidents on graphs
6. **Set up alerts** for critical metrics
7. **Export and version control** your dashboards as JSON
