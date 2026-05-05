package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryMiddleware adds tracing and metrics to HTTP requests
func TelemetryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !config.TelemetryConfig.TelemetryEnabled {
			return c.Next()
		}

		start := time.Now()

		// Extract trace context from headers
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		// Create propagation carrier from Fiber headers
		carrier := make(propagation.MapCarrier)
		c.Request().Header.VisitAll(func(key, value []byte) {
			carrier[string(key)] = string(value)
		})

		// Extract trace context using global propagator
		parentCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		// Start new span
		spanCtx, span := telemetry.Tracer.Start(parentCtx, c.Method()+" "+c.Path(),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.remote_addr", c.IP()),
			),
		)
		defer span.End()

		// Store span context in Fiber context for downstream use
		c.SetUserContext(spanCtx)

		// Continue with request
		err := c.Next()

		// Record metrics and span attributes after request
		statusCode := c.Response().StatusCode()
		duration := time.Since(start).Seconds()

		// Set span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Float64("http.duration", duration),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.RecordError(fiber.NewError(statusCode, "HTTP error"))
			if statusCode >= 500 {
				span.SetStatus(codes.Error, "Server error")
			}
		}

		// Record metrics
		httpRequestLabels := []attribute.KeyValue{
			attribute.String("method", c.Method()),
			attribute.String("route", c.Route().Path),
			attribute.String("status", strconv.Itoa(statusCode)),
		}

		if telemetry.HTTPCounter != nil {
			telemetry.HTTPCounter.Add(spanCtx, 1, metric.WithAttributes(httpRequestLabels...))
		}

		if telemetry.HTTPDuration != nil {
			telemetry.HTTPDuration.Record(spanCtx, duration, metric.WithAttributes(httpRequestLabels...))
		}

		return err
	}
}

// SpanFromContext extracts the span from Fiber context
func SpanFromContext(c *fiber.Ctx) trace.Span {
	ctx := c.UserContext()
	if ctx == nil {
		return trace.SpanFromContext(context.Background())
	}
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c *fiber.Ctx, name string, attributes ...attribute.KeyValue) {
	span := SpanFromContext(c)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(c *fiber.Ctx, attributes ...attribute.KeyValue) {
	span := SpanFromContext(c)
	span.SetAttributes(attributes...)
}
