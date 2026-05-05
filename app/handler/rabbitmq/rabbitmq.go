package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soluixdeveloper/ces-orchestratorService/app/function"
	"github.com/soluixdeveloper/ces-orchestratorService/app/helper"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	cesresponse "github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func RPCHandler(message amqp.Delivery) cesresponse.Response {
	start := time.Now()

	if message.CorrelationId == "" {
		message.CorrelationId = helper.UUIDgenerator()
	}

	ctx := context.Background()
	var span trace.Span

	// Initialize telemetry if enabled
	if config.TelemetryConfig.TelemetryEnabled && telemetry.Tracer != nil {
		// Extract trace context from message headers
		carrier := make(propagation.MapCarrier)
		if message.Headers != nil {
			for key, value := range message.Headers {
				if strValue, ok := value.(string); ok {
					carrier[key] = strValue
				}
			}
		}

		// Extract trace context using global propagator
		parentCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		// Start new span
		ctx, span = telemetry.Tracer.Start(parentCtx, "rabbitmq.rpc_handler",
			trace.WithAttributes(
				attribute.String("rabbitmq.correlation_id", message.CorrelationId),
				attribute.String("rabbitmq.routing_key", message.RoutingKey),
				attribute.String("rabbitmq.exchange", message.Exchange),
				attribute.String("rabbitmq.type", message.Type),
				attribute.Int("rabbitmq.message_size", len(message.Body)),
			),
		)
		defer span.End()
	}

	logging := ceslogger.NewLogger(message.CorrelationId)
	metadata := model.OrchestrationMetadata{}
	metadata.Header = map[string]string{}
	metadata.CorrelationID = message.CorrelationId
	metadata.RequestType = message.Type
	_ = json.Unmarshal(message.Body, &metadata)
	_ = json.Unmarshal(metadata.BodyRaw, &metadata.Body)

	type value struct {
		Data        []byte `json:"data"`
		RequestType string `json:"request_type"`
	}
	type oldModel struct {
		Key   string `json:"key"`
		Value value  `json:"value"`
	}
	oldData := oldModel{}
	_ = json.Unmarshal(message.Body, &oldData)

	if oldData.Value.Data != nil {
		_ = json.Unmarshal(oldData.Value.Data, &metadata)
		_ = json.Unmarshal(oldData.Value.Data, &metadata.Body)
	}

	// Add span attributes for workflow information
	if config.TelemetryConfig.TelemetryEnabled && span != nil {
		span.SetAttributes(
			attribute.String("workflow.request_type", metadata.RequestType),
			attribute.String("workflow.correlation_id", metadata.CorrelationID),
		)
	}

	response := function.Orchestrate(ctx, metadata, *logging)

	// Record metrics and complete span
	if config.TelemetryConfig.TelemetryEnabled {
		duration := time.Since(start).Seconds()

		// Derive status from status code
		status := "success"
		if response.StatusCode >= 400 {
			status = "error"
		}

		// Record RabbitMQ metrics
		if telemetry.RabbitMQCounter != nil {
			rabbitmqLabels := []attribute.KeyValue{
				attribute.String("type", message.Type),
				attribute.String("routing_key", message.RoutingKey),
				attribute.String("status", status),
			}
			telemetry.RabbitMQCounter.Add(ctx, 1, metric.WithAttributes(rabbitmqLabels...))
		}

		// Add span attributes for response
		if span != nil {
			span.SetAttributes(
				attribute.String("response.status", status),
				attribute.Int("response.status_code", response.StatusCode),
				attribute.Float64("processing.duration", duration),
			)

			// Set span status based on response
			if response.StatusCode >= 400 {
				span.SetStatus(codes.Error, "Processing failed")
			}
		}
	}

	return response
}
