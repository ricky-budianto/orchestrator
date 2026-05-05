# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CES-iBridge Orchestrator Service is a Go microservice that orchestrates workflows defined in YAML configurations. It listens to REST API requests and RabbitMQ RPC calls to execute complex, multi-step workflows involving workers and conditional logic.

## Key Development Commands

### Building and Running
- **Build**: `GOFLAGS="-mod=mod" go build -o orchestrator .`
- **Run locally**: `go run main.go`
- **Run tests**: `go test ./...`
- **Run specific test**: `go test ./app/helper -v`
- **Format code**: `go fmt ./...`

### Docker and Monitoring
- **Build Docker image**: `docker build -t ces-orchestrator .`
- **Run container**: `docker run -p 8081:8081 ces-orchestrator`
- **Start LGTM Stack**: `docker-compose -f docker-compose.monitoring.yml up -d` or `./start-lgtm-stack.sh`
- **Stop LGTM Stack**: `docker-compose -f docker-compose.monitoring.yml down`

## Architecture Overview

### Service Startup Flow
1. **main.go:25-50** - Initializes configurations, telemetry, Redis, and workflow definitions
2. **config/env.go:107-186** - Loads environment variables, initializes database with auto-migration, sets up RSA encryption
3. **app/function/orchestration.go:38-75** - WorkflowInit() loads all active workflow configurations from database into memory or Redis cache
4. **main.go:76-97** - Starts three concurrent services: HTTP server (Fiber), RabbitMQ consumer, and cron scheduler

### Core Workflow Execution
The orchestration engine (`app/function/orchestration.go:93+`) processes workflows in several steps:
1. **Metadata Extraction** - Captures request data (headers, body, params) from REST or RabbitMQ
2. **Workflow Lookup** - Finds workflow configuration by ID or path+method combination
3. **State Initialization** - Creates workflow_history and workflow_state database records
4. **Event Processing** - Executes startEvent → serviceTask → conditional → endEvent chain
5. **Worker Execution** - Calls external services via HTTP or RabbitMQ RPC
6. **Variable Substitution** - Replaces `${{variable}}` references with actual values from previous steps
7. **Response Assembly** - Builds final response from endEvent configuration

### Request Handling Patterns

**V1 API (Static)**: `/api/ibridge/v1/*` routes are registered at startup based on `path` and `method` in workflow_configurations table. Changes require service restart.

**V2 API (Dynamic)**: `/api/ibridge/v2/orchestrate/:workflow_configuration_id` provides instant workflow execution using the `id` field as lookup key. No restart needed for new workflows.

**RabbitMQ RPC**: Messages with `request_type` field are matched against workflow configurations. Queue name is `ces_orchestrator` (configurable via RABBITMQURL).

### Database Schema
- **workflow_configurations**: Stores YAML workflow definitions (base64 encoded in `configuration` column). Fields `path` and `method` override YAML values.
- **workflow_histories**: Audit trail of executions with request/response and status (ACTIVE/EXECUTED/FAILED).
- **workflow_states**: Detailed per-worker state for each execution, keyed by `workflow_id` (references workflow_histories.id).
- **workflow_audits**: Change tracking for workflow configuration updates.

### Variable Substitution System
The engine supports complex variable references in workflow YAML:
- `${{startEvent.Body.email}}` - Access request body fields
- `${{worker_name.data.field}}` - Reference previous worker outputs
- `${{ibridge.httprequest}}` - Special parsing for HTTP request format
- Global state stored in `workflowState` map, accessible across all workers in execution

### Telemetry and Observability

**LGTM Stack** (Loki, Grafana, Tempo, Prometheus):
- **Tempo** receives OTLP traces via HTTP (port 4318) or gRPC (port 4317)
- **Prometheus** scrapes `/metrics` endpoint on port 8081 (or APP_PORT)
- **Loki** aggregates logs (port 3100)
- **Grafana** provides unified visualization (port 3001, credentials: admin/admin)

**Instrumentation**:
- HTTP requests traced via `app/middleware/telemetry.go` (TelemetryMiddleware)
- Workflow executions create spans in `app/function/orchestration.go:93-100`
- Database operations automatically traced via GORM callbacks in `config/telemetry/gorm_tracing.go`
  - All database queries (SELECT, INSERT, UPDATE, DELETE) create child spans
  - Span attributes include: db.system, db.name, db.operation, db.table, db.statement, db.rows_affected
  - Error handling: Failed queries recorded with error events and error status
  - Context propagation: Requires using `db.WithContext(ctx)` in DAL functions
- Custom metrics: `orchestrator_workflows_total`, `orchestrator_http_requests_total`, `orchestrator_workflow_duration_seconds`

**Configuration**:
- Enable with `TELEMETRY_ENABLED=true`
- Set `TEMPO_ENDPOINT=localhost:4318` (OTLP HTTP endpoint without protocol prefix)
- Metrics automatically exposed at `/metrics` endpoint

**Querying Database Traces in Grafana Tempo**:

Access Grafana at `http://localhost:3001` (admin/admin) and use these queries in the Explore view:

```
# Find all database operations
{span.db.system="postgresql"}

# Find specific operation types
{span.db.operation="INSERT"}
{span.db.operation="SELECT"}

# Find operations on specific tables
{span.db.table="workflow_histories"}
{span.db.table="workflow_states"}

# Combine filters
{span.db.system="postgresql" && span.db.operation="INSERT" && span.db.table="workflow_histories"}
```

**Expected Span Hierarchy for Workflow Execution**:
```
HTTP POST /api/ibridge/v1/workflow-path (parent)
├── workflow: execute workflow_name (child)
│   ├── db.query: query workflow_configurations (grandchild)
│   ├── db.query: create workflow_histories (grandchild)
│   ├── worker: call_external_service (grandchild)
│   ├── db.query: create workflow_states (grandchild)
│   └── db.query: update workflow_histories (grandchild)
```

For comprehensive E2E testing guide, see `.spec-workflow/specs/database-tracing/E2E_TESTING_GUIDE.md`

## Critical Configuration

### Environment Variables
- `TEMPO_ENDPOINT`: OTLP endpoint for traces (e.g., `localhost:4318`)
- `TELEMETRY_ENABLED`: Enable/disable observability stack (true/false)
- `POSTGRE_SQL_*`: Database connection (auto-migration on startup)
- `RABBITMQURL`: RabbitMQ connection string
- `REDIS_*`: Caching layer (if `REDIS_USE=true`)
- `CRYPTO_RSA`: Base64-encoded RSA private key for encryption
- `APP_PORT`: HTTP server port (default 8081)

### Workflow YAML Structure

**See `WORKFLOW_GUIDE.md` for comprehensive workflow documentation.**

Basic structure:
```yaml
name: workflow_id              # Must match database id field
requestType: request_type      # For RabbitMQ routing
path: /api/ibridge/v1/path     # REST endpoint (overridden by DB)
method: POST                   # HTTP method (overridden by DB)
startEvent:
  targetRef: first_task        # Entry point

serviceTask:
  task_name:
    type: worker|conditional|endEvent
    sourceRef: {...}           # Input data with variable refs
    targetRef: [next_task]     # Routing to next task(s)
    worker: middleware         # Worker identifier for RabbitMQ
```

**Variable Substitution:**
- `${{startEvent.Body.field}}` - Request body
- `${{worker_name.data.field}}` - Previous worker results
- `${{ibridge.global.field}}` - Global variables
- `${{ibridge.timestamp}}` - Current timestamp

## Development Guidelines

### Adding New Workflows

**For detailed workflow syntax and examples, see `WORKFLOW_GUIDE.md`**

Quick steps:
1. Create YAML configuration with unique `name`
2. POST to `/api/ibridge/v1/workflow_configuration` with base64-encoded YAML in `configuration` field
3. Set `path` and `method` in request body (these override YAML values)
4. For instant activation, use V2 endpoint: `/api/ibridge/v2/orchestrate/:workflow_id`
5. For traditional routing, restart service to register new route

Sample workflows available in `sample/` directory.

### API Versioning Rules
- **Never use `/v2` in workflow path configurations** - reserved for dynamic routing
- V1 paths require service restart after changes
- V2 uses `workflow_configuration_id` parameter, no restart needed
- Database `path` and `method` fields take precedence over YAML

### Testing Workflows
1. Check workflow syntax with YAML validator
2. Test variable substitution with sample data
3. Monitor execution in `workflow_histories` table
4. Debug worker states in `workflow_states` table
5. View traces in Grafana Tempo for performance analysis

### Graceful Shutdown
Service handles SIGTERM/SIGINT signals:
1. Stops accepting new HTTP requests
2. Cancels context for RabbitMQ consumer
3. Stops cron scheduler
4. Flushes telemetry data (10s timeout)
5. Closes database connections
6. 30-second grace period before forced shutdown

### Database Auto-Migration
Tables auto-create/update on startup via GORM at `config/env.go:172-178`. Models are in `app/model/`. Use GORM conventions (e.g., `ID string` becomes `id` column).

### Logging
- Custom CES logger via `ceslogger.Logger{}`
- Elasticsearch integration if configured
- Daily cleanup job at 2:00 AM removes old workflow_history and workflow_state records based on `RETENTION_LOG` days

### Dependencies
- **Fiber v2**: Web framework
- **GORM**: PostgreSQL ORM
- **RabbitMQ AMQP**: Message queue (ces-utilities wrapper)
- **Redis**: Optional caching layer
- **OpenTelemetry**: Observability (OTLP exporters)
- **ces-utilities**: Internal libraries (logger, response, rabbitmq helpers)

## Important Notes

### Service Port
The service runs on `APP_PORT` (default 8081, not 3000). Update monitoring configs and deployment manifests accordingly.

### Crypto Operations
RSA encryption/decryption uses `CRYPTO_RSA` environment variable. Format: base64-encoded PKCS1 private key without headers. See `config/env.go:188-219` for implementation.

### Workflow Execution Tracking
Every workflow execution creates:
- One `workflow_histories` record (request/response/status)
- One `workflow_states` record (per-worker state details)
- OpenTelemetry trace with spans for each worker (if telemetry enabled)

### RabbitMQ Message Format
```json
{
  "request_id": "unique_id",
  "correlation_id": "correlation_id",
  "request_type": "workflow_name",
  "value": { "request": "data" },
  "need_response": true
}
```

### Deployment
Containerized with distroless base image. Kubernetes manifests in README show ConfigMap, Secret, Deployment, Service, and Ingress examples. GitHub Actions deploys to AWS ECR.

### Performance Considerations
- Workflow configs cached in Redis (if enabled) or in-memory map
- Concurrent worker execution within single workflow not yet supported
- Database connection pooling configured via `POSTGRE_SQL_MAX_OPEN_CONNECTION` and `POSTGRE_SQL_MAX_IDLE_CONNECTION`
