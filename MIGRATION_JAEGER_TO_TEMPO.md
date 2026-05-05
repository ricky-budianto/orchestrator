# Migration from Jaeger to Tempo - LGTM Stack Implementation

## Overview
Successfully migrated the CES Orchestrator Service from using Jaeger for distributed tracing to using Grafana Tempo as part of the LGTM (Loki, Grafana, Tempo, Mimir) observability stack.

## Changes Made

### 1. OpenTelemetry Configuration Updates

#### config/telemetry/telemetry.go
- **Changed**: Replaced Jaeger exporter with OTLP HTTP exporter
- **Before**: `go.opentelemetry.io/otel/exporters/jaeger`
- **After**: `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
- **Config Struct**: Updated `JaegerEndpoint` to `TempoEndpoint`
- **Endpoint Format**: Changed from HTTP collector URL to OTLP endpoint (e.g., `localhost:4318`)

#### config/env.go
- Updated `EnvConfigTelemetry` struct:
  - Renamed field from `JaegerEndpoint` to `TempoEndpoint`
  - Maps to env var `TEMPO_ENDPOINT`

#### main.go
- Updated telemetry initialization to use `TempoEndpoint` instead of `JaegerEndpoint`

### 2. Environment Configuration

#### config.env.example
```env
# Before
JAEGER_ENDPOINT=http://localhost:14268/api/traces

# After
TEMPO_ENDPOINT=localhost:4318
```

**Note**: Update your `config.env` file with the new `TEMPO_ENDPOINT` variable.

### 3. LGTM Stack Docker Compose

#### docker-compose.monitoring.yml
Complete rewrite to include full LGTM stack:

**Services**:
- **Tempo** (Port 3200, 4317 gRPC, 4318 HTTP) - Distributed tracing backend
- **Loki** (Port 3100) - Log aggregation system
- **Prometheus** (Port 9090) - Metrics collection and storage
- **Grafana** (Port 3001) - Unified visualization dashboard

**Volumes**:
- `tempo-data` - Tempo trace storage
- `loki-data` - Loki log storage
- `prometheus-data` - Prometheus metrics storage
- `grafana-storage` - Grafana configuration

### 4. Configuration Files Created

#### tempo.yaml
- Configured OTLP receivers for HTTP (port 4318) and gRPC (port 4317)
- Set up local storage backend for traces
- Enabled metrics generator with Prometheus remote write
- Configured compaction and retention policies

#### grafana-datasources.yaml
Automatic datasource provisioning for:
- **Prometheus**: Default datasource with exemplar support
- **Tempo**: Configured with:
  - Trace-to-logs correlation (Loki)
  - Trace-to-metrics correlation (Prometheus)
  - Service map visualization
  - Node graph support
- **Loki**: Configured with trace ID extraction from logs

#### prometheus.yml
Updated scrape configs:
- Added Tempo metrics scraping (port 3200)
- Updated orchestrator service target to port 8081
- Maintained existing Prometheus self-monitoring

### 5. Go Dependencies

#### go.mod Changes
```diff
- go.opentelemetry.io/otel/exporters/jaeger v1.17.0
+ go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
```

**Upgraded packages**:
- OpenTelemetry core libraries: v1.37.0 → v1.38.0
- Go version: 1.23.0 → 1.24.0
- Added OTLP proto dependencies
- Updated genproto packages to split modules

## How to Use

### 1. Update Environment Variables
Copy the new configuration from `config.env.example`:
```bash
# Update your config.env file
TEMPO_ENDPOINT=localhost:4318
```

### 2. Start the LGTM Stack
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### 3. Build and Run the Service
```bash
# Build the application
GOFLAGS="-mod=mod" go build -o orchestrator .

# Run the service
./orchestrator
```

### 4. Access the UIs
- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200
- **Loki**: http://localhost:3100

### 5. View Traces in Grafana
1. Navigate to Grafana at http://localhost:3001
2. Go to Explore → Select "Tempo" datasource
3. Use TraceQL or search to find traces
4. Click on traces to see:
   - Trace timeline and spans
   - Related logs (from Loki)
   - Related metrics (from Prometheus)
   - Service graph visualization

## Benefits of LGTM Stack

### 1. **Unified Observability**
- Logs (Loki) + Metrics (Prometheus) + Traces (Tempo) in one stack
- Seamless correlation between signals
- Jump from traces to logs to metrics

### 2. **Cost-Effective**
- Tempo uses object storage (cheaper than databases)
- Loki uses compressed indexes (lower storage costs)
- All components are open-source

### 3. **Scalability**
- Tempo designed for high-volume tracing
- Horizontal scaling support
- Efficient compression and storage

### 4. **Better Visualization**
- TraceQL query language
- Service graph visualization
- Node graph for service dependencies
- Exemplar support linking metrics to traces

### 5. **Vendor Neutrality**
- Standards-based (OTLP, Prometheus)
- No vendor lock-in
- Compatible with existing tools

## Differences from Jaeger

| Feature | Jaeger | Tempo |
|---------|--------|-------|
| Protocol | Jaeger native + OTLP | OTLP only |
| Storage | Cassandra/Elasticsearch | Object storage (S3-compatible) |
| UI | Built-in UI | Grafana (more powerful) |
| Search | Full-text search | TraceQL + metadata search |
| Metrics | Basic | Metrics generator + exemplars |
| Logs | No integration | Native Loki integration |
| Cost | Higher (storage) | Lower (object storage) |

## Troubleshooting

### Traces Not Appearing
1. Check Tempo logs: `docker-compose logs tempo`
2. Verify TEMPO_ENDPOINT in config.env
3. Ensure Tempo is receiving data: Check http://localhost:3200/status
4. Verify telemetry is enabled: `TELEMETRY_ENABLED=true`

### Build Issues
If you encounter `missing go.sum entry` errors:
```bash
GOFLAGS="-mod=mod" go build -o orchestrator .
```

### Grafana Datasources Not Loading
1. Check datasource configuration: `docker-compose logs grafana`
2. Restart Grafana: `docker-compose restart grafana`
3. Manually add datasources via Grafana UI if needed

### Tempo Connection Issues
Ensure Tempo is healthy:
```bash
curl http://localhost:3200/ready
curl http://localhost:3200/status
```

## Migration Checklist

- [x] Update Go code to use OTLP exporter
- [x] Update environment configuration
- [x] Create LGTM stack docker-compose
- [x] Configure Tempo for OTLP ingestion
- [x] Set up Grafana datasources
- [x] Update Prometheus scrape configs
- [x] Build and test application
- [ ] Update production config.env
- [ ] Deploy LGTM stack to production
- [ ] Verify traces are flowing
- [ ] Update monitoring dashboards
- [ ] Document for team

## Next Steps

1. **Create Grafana Dashboards**
   - Import pre-built Tempo dashboards
   - Create custom dashboards for orchestrator metrics
   - Set up alerts for trace errors

2. **Configure Log Shipping**
   - Send application logs to Loki
   - Correlate logs with traces using trace IDs
   - Set up log-based alerts

3. **Production Deployment**
   - Use external storage for Tempo (MinIO/S3)
   - Set up persistent volumes for all services
   - Configure authentication and authorization
   - Enable HTTPS/TLS for all components

4. **Optimize Sampling**
   - Implement adaptive sampling
   - Configure tail-based sampling
   - Reduce storage costs while maintaining visibility

## Resources

- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [TraceQL Documentation](https://grafana.com/docs/tempo/latest/traceql/)
- [LGTM Stack Guide](https://grafana.com/blog/2021/05/21/lgtm-stack/)
