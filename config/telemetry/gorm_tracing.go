package telemetry

import (
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// Span context key for storing spans in GORM instance values
const spanContextKey = "otel:span"

// RegisterGormTracing registers OpenTelemetry tracing callbacks for GORM database operations
// Note: This function should be called after telemetry is initialized
func RegisterGormTracing(db *gorm.DB, dbName string) error {
	logging := ceslogger.Logger{}

	// Verify tracer is initialized
	if Tracer == nil {
		logging.LogInfo("GORM tracing skipped - tracer not initialized")
		return nil
	}

	// Register create callbacks
	if err := db.Callback().Create().Before("create:before").Register("otel:before_create", beforeCallback("create")); err != nil {
		logging.LogError("Failed to register before_create callback", err)
		return err
	}
	if err := db.Callback().Create().After("create:after").Register("otel:after_create", afterCallback("create", dbName)); err != nil {
		logging.LogError("Failed to register after_create callback", err)
		return err
	}

	// Register query callbacks
	if err := db.Callback().Query().Before("query:before").Register("otel:before_query", beforeCallback("query")); err != nil {
		logging.LogError("Failed to register before_query callback", err)
		return err
	}
	if err := db.Callback().Query().After("query:after").Register("otel:after_query", afterCallback("query", dbName)); err != nil {
		logging.LogError("Failed to register after_query callback", err)
		return err
	}

	// Register update callbacks
	if err := db.Callback().Update().Before("update:before").Register("otel:before_update", beforeCallback("update")); err != nil {
		logging.LogError("Failed to register before_update callback", err)
		return err
	}
	if err := db.Callback().Update().After("update:after").Register("otel:after_update", afterCallback("update", dbName)); err != nil {
		logging.LogError("Failed to register after_update callback", err)
		return err
	}

	// Register delete callbacks
	if err := db.Callback().Delete().Before("delete:before").Register("otel:before_delete", beforeCallback("delete")); err != nil {
		logging.LogError("Failed to register before_delete callback", err)
		return err
	}
	if err := db.Callback().Delete().After("delete:after").Register("otel:after_delete", afterCallback("delete", dbName)); err != nil {
		logging.LogError("Failed to register after_delete callback", err)
		return err
	}

	logging.LogInfo("GORM tracing callbacks registered successfully")
	return nil
}

// beforeCallback creates a span before database operation
func beforeCallback(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		// Extract context from GORM statement
		ctx := db.Statement.Context
		if ctx == nil {
			return
		}

		// Check if there's a valid trace context
		if trace.SpanFromContext(ctx).SpanContext().TraceID().IsValid() {
			// Extract table name for span name
			tableName := extractTableName(db)
			spanName := "db.query: " + operation + " " + tableName

			// Start new span
			_, span := Tracer.Start(ctx, spanName)

			// Store span in GORM instance for afterCallback
			db.InstanceSet(spanContextKey, span)
		}
	}
}

// afterCallback enriches and ends the span after database operation
func afterCallback(operation string, dbName string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		// Retrieve span from instance
		spanVal, exists := db.InstanceGet(spanContextKey)
		if !exists {
			return
		}

		span, ok := spanVal.(trace.Span)
		if !ok {
			return
		}
		defer span.End()

		// Add database operation attributes
		attrs := []attribute.KeyValue{
			attribute.String("db.system", "postgresql"),
			attribute.String("db.name", dbName),
			attribute.String("db.operation", extractOperation(operation)),
			attribute.String("db.table", extractTableName(db)),
		}

		// Add SQL statement if available
		if db.Statement.SQL.String() != "" {
			attrs = append(attrs, attribute.String("db.statement", db.Statement.SQL.String()))
		}

		// Add rows affected for write operations
		if operation != "query" && db.Statement.RowsAffected > 0 {
			attrs = append(attrs, attribute.Int64("db.rows_affected", db.Statement.RowsAffected))
		}

		span.SetAttributes(attrs...)

		// Set error status if query failed
		if db.Statement.Error != nil {
			span.RecordError(db.Statement.Error)
			span.SetStatus(codes.Error, db.Statement.Error.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

// extractTableName extracts the table name from GORM statement
func extractTableName(db *gorm.DB) string {
	if db.Statement.Table != "" {
		return db.Statement.Table
	}

	// Try to extract from schema
	if db.Statement.Schema != nil && db.Statement.Schema.Table != "" {
		return db.Statement.Schema.Table
	}

	// Fallback to unknown
	return "unknown"
}

// extractOperation converts operation name to uppercase SQL command
func extractOperation(operation string) string {
	switch operation {
	case "create":
		return "INSERT"
	case "query":
		return "SELECT"
	case "update":
		return "UPDATE"
	case "delete":
		return "DELETE"
	default:
		return operation
	}
}
