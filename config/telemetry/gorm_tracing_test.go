package telemetry

import (
	"context"
	"testing"

	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestExtractTableName tests the extractTableName function
func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *gorm.DB
		expected string
	}{
		{
			name: "table name from Statement.Table",
			setup: func() *gorm.DB {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				db.Statement = &gorm.Statement{Table: "users"}
				return db
			},
			expected: "users",
		},
		{
			name: "table name with Schema but no Table",
			setup: func() *gorm.DB {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				// Create a test model to get schema
				type TestModel struct {
					ID uint
				}
				stmt := &gorm.Statement{DB: db}
				stmt.Parse(&TestModel{})
				db.Statement = stmt
				return db
			},
			expected: "test_models",
		},
		{
			name: "fallback to unknown",
			setup: func() *gorm.DB {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				db.Statement = &gorm.Statement{}
				return db
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			result := extractTableName(db)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractOperation tests the extractOperation function
func TestExtractOperation(t *testing.T) {
	tests := []struct {
		operation string
		expected  string
	}{
		{"create", "INSERT"},
		{"query", "SELECT"},
		{"update", "UPDATE"},
		{"delete", "DELETE"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			result := extractOperation(tt.operation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegisterGormTracing_TelemetryDisabled tests that registration is skipped when tracer is nil
func TestRegisterGormTracing_TracerNotInitialized(t *testing.T) {
	// Initialize logger to avoid nil pointer
	ceslogger.InitLogger("0", "test", "test", "test", "1.0")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Save original tracer and set to nil
	originalTracer := Tracer
	Tracer = nil
	defer func() { Tracer = originalTracer }()

	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err) // Should not error, just skip
}

// TestRegisterGormTracing_Success tests successful callback registration
func TestRegisterGormTracing_Success(t *testing.T) {
	// Initialize logger
	ceslogger.InitLogger("0", "test", "test", "test", "1.0")

	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	Tracer = tp.Tracer("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err) // Should succeed
}

// TestGormTracing_CreateOperation tests span creation for INSERT operations
func TestGormTracing_CreateOperation(t *testing.T) {
	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	Tracer = tp.Tracer("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test table
	type TestModel struct {
		ID   uint
		Name string
	}
	db.AutoMigrate(&TestModel{})

	// Register tracing
	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err)

	// Create parent span context
	ctx, parentSpan := Tracer.Start(context.Background(), "test-parent")
	defer parentSpan.End()

	// Execute create operation with context
	result := db.WithContext(ctx).Create(&TestModel{Name: "test"})
	assert.NoError(t, result.Error)

	// Force flush spans
	tp.ForceFlush(context.Background())

	// Verify spans were created
	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 1, "At least one span should be created")

	// Verify database span exists with correct attributes
	foundDBSpan := false
	for i := range spans {
		attrs := spans[i].Attributes
		attrMap := make(map[string]interface{})
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsInterface()
		}

		// Check if this is a database span
		if attrMap["db.system"] == "postgresql" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			assert.Equal(t, "INSERT", attrMap["db.operation"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database span should be created")
}

// TestGormTracing_QueryOperation tests span creation for SELECT operations
func TestGormTracing_QueryOperation(t *testing.T) {
	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	Tracer = tp.Tracer("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test table
	type TestModel struct {
		ID   uint
		Name string
	}
	db.AutoMigrate(&TestModel{})
	db.Create(&TestModel{Name: "test"})

	// Register tracing
	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err)

	// Clear previous spans
	exporter.Reset()

	// Create parent span context
	ctx, parentSpan := Tracer.Start(context.Background(), "test-parent")
	defer parentSpan.End()

	// Execute query operation with context
	var results []TestModel
	result := db.WithContext(ctx).Find(&results)
	assert.NoError(t, result.Error)

	// Force flush spans
	tp.ForceFlush(context.Background())

	// Verify spans were created
	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 1, "At least one span should be created")

	// Verify database span exists
	foundDBSpan := false
	for i := range spans {
		attrs := spans[i].Attributes
		attrMap := make(map[string]interface{})
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsInterface()
		}

		// Check if this is a database span
		if attrMap["db.system"] == "postgresql" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			assert.Equal(t, "SELECT", attrMap["db.operation"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database span should be created")
}

// TestGormTracing_ErrorHandling tests that errors are recorded in spans
func TestGormTracing_ErrorHandling(t *testing.T) {
	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	Tracer = tp.Tracer("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Register tracing
	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err)

	// Create parent span context
	ctx, parentSpan := Tracer.Start(context.Background(), "test-parent")
	defer parentSpan.End()

	// Execute query on non-existent table (will error)
	var results []map[string]interface{}
	db.WithContext(ctx).Table("non_existent_table").Find(&results)

	// Force flush spans
	tp.ForceFlush(context.Background())

	// Verify spans were created
	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 1, "At least one span should be created")

	// Find database span with error
	foundErrorSpan := false
	for i := range spans {
		attrs := spans[i].Attributes
		attrMap := make(map[string]interface{})
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsInterface()
		}

		// Check if this is a database span
		if attrMap["db.system"] == "postgresql" {
			foundErrorSpan = true
			// Verify span has error event
			assert.Greater(t, len(spans[i].Events), 0, "Span should have error event recorded")
			break
		}
	}

	assert.True(t, foundErrorSpan, "Database span should be created even with errors")
}

// TestGormTracing_NoContextGracefulHandling tests that operations without context don't panic
func TestGormTracing_NoContextGracefulHandling(t *testing.T) {
	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	Tracer = tp.Tracer("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test table
	type TestModel struct {
		ID   uint
		Name string
	}
	db.AutoMigrate(&TestModel{})

	// Register tracing
	err = RegisterGormTracing(db, "test_db")
	assert.NoError(t, err)

	// Execute operation WITHOUT context - should not panic
	assert.NotPanics(t, func() {
		db.Create(&TestModel{Name: "test"})
	})
}
