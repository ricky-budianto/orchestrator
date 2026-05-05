package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"net/http"
)

var (
	Tracer           oteltrace.Tracer
	Meter            metric.Meter
	MetricsHandler   http.Handler
	TelemetryCleanup func(context.Context) error
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	TempoEndpoint  string
	Environment    string
}

func InitTelemetry(cfg Config) error {
	logging := ceslogger.Logger{}

	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		logging.LogError("failed to create telemetry resource", err.Error())
		return fmt.Errorf("failed to create telemetry resource: %w", err)
	}

	// Initialize Tempo tracing via OTLP
	if err := initTracing(ctx, res, cfg.TempoEndpoint, &logging); err != nil {
		logging.LogError("failed to initialize tracing", err.Error())
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize Prometheus metrics
	if err := initMetrics(ctx, res, &logging); err != nil {
		logging.LogError("failed to initialize metrics", err.Error())
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Set up text map propagator for distributed tracing (W3C TraceContext + Baggage)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer and meter instances
	Tracer = otel.Tracer(cfg.ServiceName)
	Meter = otel.Meter(cfg.ServiceName)

	logging.LogInfo("Telemetry initialized successfully")
	return nil
}

func initTracing(ctx context.Context, res *resource.Resource, tempoEndpoint string, logging *ceslogger.Logger) error {
	// Create OTLP HTTP exporter for Tempo
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(tempoEndpoint),
		otlptracehttp.WithInsecure(), // Use WithTLSCredentials() for production
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tp)

	// Set cleanup function for graceful shutdown
	TelemetryCleanup = func(ctx context.Context) error {
		// Create a context with timeout for shutdown
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return tp.Shutdown(shutdownCtx)
	}

	logging.LogInfo("Tracing initialized with Tempo endpoint: " + tempoEndpoint)
	return nil
}

func initMetrics(ctx context.Context, res *resource.Resource, logging *ceslogger.Logger) error {
	// Create Prometheus registry
	registry := prometheus.NewRegistry()

	// Create Prometheus exporter
	exporter, err := promexporter.New(
		promexporter.WithRegisterer(registry),
		promexporter.WithoutTargetInfo(),
	)
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	// Set the global meter provider
	otel.SetMeterProvider(mp)

	// Create HTTP handler for metrics endpoint
	MetricsHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	})

	logging.LogInfo("Metrics initialized with Prometheus exporter")
	return nil
}

// CreateCustomMetrics creates common metrics for the orchestrator service
func CreateCustomMetrics() error {
	logging := ceslogger.Logger{}

	// Workflow execution counter
	workflowCounter, err := Meter.Int64Counter(
		"orchestrator_workflows_total",
		metric.WithDescription("Total number of workflows executed"),
	)
	if err != nil {
		logging.LogError("failed to create workflow counter", err.Error())
		return err
	}

	// Workflow duration histogram
	workflowDuration, err := Meter.Float64Histogram(
		"orchestrator_workflow_duration_seconds",
		metric.WithDescription("Duration of workflow execution in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logging.LogError("failed to create workflow duration histogram", err.Error())
		return err
	}

	// HTTP request counter
	httpCounter, err := Meter.Int64Counter(
		"orchestrator_http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		logging.LogError("failed to create HTTP counter", err.Error())
		return err
	}

	// HTTP request duration histogram
	httpDuration, err := Meter.Float64Histogram(
		"orchestrator_http_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logging.LogError("failed to create HTTP duration histogram", err.Error())
		return err
	}

	// RabbitMQ message counter
	rabbitmqCounter, err := Meter.Int64Counter(
		"orchestrator_rabbitmq_messages_total",
		metric.WithDescription("Total number of RabbitMQ messages processed"),
	)
	if err != nil {
		logging.LogError("failed to create RabbitMQ counter", err.Error())
		return err
	}

	// Store metrics globally for use in handlers
	WorkflowCounter = workflowCounter
	WorkflowDuration = workflowDuration
	HTTPCounter = httpCounter
	HTTPDuration = httpDuration
	RabbitMQCounter = rabbitmqCounter

	logging.LogInfo("Custom metrics created successfully")
	return nil
}

// Global metrics instances
var (
	WorkflowCounter  metric.Int64Counter
	WorkflowDuration metric.Float64Histogram
	HTTPCounter      metric.Int64Counter
	HTTPDuration     metric.Float64Histogram
	RabbitMQCounter  metric.Int64Counter
)