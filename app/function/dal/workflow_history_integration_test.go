package dal

import (
	"context"
	"testing"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database with telemetry for testing
func setupTestDB(t *testing.T) (*gorm.DB, *tracetest.InMemoryExporter) {
	// Initialize logger
	ceslogger.InitLogger("0", "test", "test", "test", "1.0")

	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Initialize global tracer
	telemetry.Tracer = tp.Tracer("test")

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate test tables (only WorkflowHistory and WorkflowConfiguration to avoid PostgreSQL-specific syntax)
	err = db.AutoMigrate(
		&model.WorkflowHistory{},
		&model.WorkflowConfiguration{},
	)
	assert.NoError(t, err)

	// Register GORM tracing
	err = telemetry.RegisterGormTracing(db, "test_db")
	assert.NoError(t, err)

	// Set as global database for DAL functions
	config.PostgreDB = db

	return db, exporter
}

// TestCreateWorkflowHistory_WithTracing tests that CreateWorkflowHistory generates database spans
func TestCreateWorkflowHistory_WithTracing(t *testing.T) {
	db, exporter := setupTestDB(t)
	defer func() { config.PostgreDB = nil }()

	// Create parent span context
	ctx, parentSpan := telemetry.Tracer.Start(context.Background(), "test-create-workflow-history")
	defer parentSpan.End()

	// Create test data with context
	data := &model.WorkflowHistory{
		Id:                      "test-history-id",
		WorkflowConfigurationId: "test-workflow-id",
		Request:                 map[string]interface{}{"test": "data"},
		Response:                map[string]interface{}{"result": "success"},
		Status:                  "EXECUTED",
	}

	// Execute create operation with context
	err := db.WithContext(ctx).Create(data).Error
	assert.NoError(t, err)

	// Force flush spans
	otel.GetTracerProvider().(*trace.TracerProvider).ForceFlush(context.Background())

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

		// Check if this is a database span for INSERT operation
		if attrMap["db.system"] == "postgresql" && attrMap["db.operation"] == "INSERT" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			assert.Equal(t, "workflow_histories", attrMap["db.table"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database INSERT span should be created")
}

// TestGetWorkflowHistory_WithTracing tests that GetWorkflowHistory generates database spans
func TestGetWorkflowHistory_WithTracing(t *testing.T) {
	db, exporter := setupTestDB(t)
	defer func() { config.PostgreDB = nil }()

	// Create test data first
	data := &model.WorkflowHistory{
		Id:                      "test-history-id-2",
		WorkflowConfigurationId: "test-workflow-id-2",
		Status:                  "ACTIVE",
	}
	err := db.Create(data).Error
	assert.NoError(t, err)

	// Clear previous spans
	exporter.Reset()

	// Create parent span context
	ctx, parentSpan := telemetry.Tracer.Start(context.Background(), "test-get-workflow-history")
	defer parentSpan.End()

	// Execute get operation with context
	var result model.WorkflowHistory
	result.Id = "test-history-id-2"
	err = db.WithContext(ctx).First(&result).Error
	assert.NoError(t, err)

	// Force flush spans
	otel.GetTracerProvider().(*trace.TracerProvider).ForceFlush(context.Background())

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

		// Check if this is a database span for SELECT operation
		if attrMap["db.system"] == "postgresql" && attrMap["db.operation"] == "SELECT" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database SELECT span should be created")
}

// TestUpdateWorkflowHistory_WithTracing tests that UpdateWorkflowHistory generates database spans
func TestUpdateWorkflowHistory_WithTracing(t *testing.T) {
	db, exporter := setupTestDB(t)
	defer func() { config.PostgreDB = nil }()

	// Create test data first
	data := &model.WorkflowHistory{
		Id:                      "test-history-id-3",
		WorkflowConfigurationId: "test-workflow-id-3",
		Status:                  "ACTIVE",
	}
	err := db.Create(data).Error
	assert.NoError(t, err)

	// Clear previous spans
	exporter.Reset()

	// Create parent span context
	ctx, parentSpan := telemetry.Tracer.Start(context.Background(), "test-update-workflow-history")
	defer parentSpan.End()

	// Execute update operation with context
	updateData := &model.WorkflowHistory{
		Status: "EXECUTED",
	}
	err = db.WithContext(ctx).Model(&model.WorkflowHistory{}).Where("id = ?", "test-history-id-3").Updates(updateData).Error
	assert.NoError(t, err)

	// Force flush spans
	otel.GetTracerProvider().(*trace.TracerProvider).ForceFlush(context.Background())

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

		// Check if this is a database span for UPDATE operation
		if attrMap["db.system"] == "postgresql" && attrMap["db.operation"] == "UPDATE" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			// Verify rows affected attribute exists
			assert.NotNil(t, attrMap["db.rows_affected"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database UPDATE span should be created")
}

// TestDeleteWorkflowHistory_WithTracing tests that DeleteWorkflowHistory generates database spans
func TestDeleteWorkflowHistory_WithTracing(t *testing.T) {
	db, exporter := setupTestDB(t)
	defer func() { config.PostgreDB = nil }()

	// Create test data first
	data := &model.WorkflowHistory{
		Id:                      "test-history-id-4",
		WorkflowConfigurationId: "test-workflow-id-4",
		Status:                  "FAILED",
	}
	err := db.Create(data).Error
	assert.NoError(t, err)

	// Clear previous spans
	exporter.Reset()

	// Create parent span context
	ctx, parentSpan := telemetry.Tracer.Start(context.Background(), "test-delete-workflow-history")
	defer parentSpan.End()

	// Execute delete operation with context
	err = db.WithContext(ctx).Unscoped().Where("id = ?", "test-history-id-4").Delete(&model.WorkflowHistory{}).Error
	assert.NoError(t, err)

	// Force flush spans
	otel.GetTracerProvider().(*trace.TracerProvider).ForceFlush(context.Background())

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

		// Check if this is a database span for DELETE operation
		if attrMap["db.system"] == "postgresql" && attrMap["db.operation"] == "DELETE" {
			foundDBSpan = true
			assert.Equal(t, "test_db", attrMap["db.name"])
			assert.Equal(t, "workflow_histories", attrMap["db.table"])
			break
		}
	}

	assert.True(t, foundDBSpan, "Database DELETE span should be created")
}

// TestListWorkflowHistory_WithTracing tests that ListWorkflowHistory with complex query generates spans
func TestListWorkflowHistory_WithTracing(t *testing.T) {
	db, exporter := setupTestDB(t)
	defer func() { config.PostgreDB = nil }()

	// Create multiple test records
	for i := 1; i <= 5; i++ {
		data := &model.WorkflowHistory{
			Id:                      "test-history-" + string(rune('a'+i-1)),
			WorkflowConfigurationId: "workflow-1",
			Status:                  "EXECUTED",
		}
		err := db.Create(data).Error
		assert.NoError(t, err)
	}

	// Clear previous spans
	exporter.Reset()

	// Create parent span context
	ctx, parentSpan := telemetry.Tracer.Start(context.Background(), "test-list-workflow-history")
	defer parentSpan.End()

	// Execute list query with pagination and filtering (with context)
	var results []model.WorkflowHistory
	var totalData int64
	err := db.WithContext(ctx).Model(&model.WorkflowHistory{}).
		Where("status = ?", "EXECUTED").
		Count(&totalData).
		Limit(10).
		Offset(0).
		Find(&results).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 5)

	// Force flush spans
	otel.GetTracerProvider().(*trace.TracerProvider).ForceFlush(context.Background())

	// Verify spans were created
	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 1, "At least one span should be created")

	// Verify database span exists (should have at least 2 SELECT operations: Count + Find)
	selectSpanCount := 0
	for i := range spans {
		attrs := spans[i].Attributes
		attrMap := make(map[string]interface{})
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsInterface()
		}

		// Check if this is a database span for SELECT operation
		if attrMap["db.system"] == "postgresql" && attrMap["db.operation"] == "SELECT" {
			selectSpanCount++
			assert.Equal(t, "test_db", attrMap["db.name"])
			assert.Equal(t, "workflow_histories", attrMap["db.table"])
		}
	}

	assert.GreaterOrEqual(t, selectSpanCount, 2, "At least 2 SELECT spans should be created (Count + Find)")
}
