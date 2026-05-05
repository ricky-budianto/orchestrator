package telemetry

import (
	"context"
	"fmt"
	"testing"

	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupBenchmarkDB creates a test database for benchmarking
func setupBenchmarkDB(enableTracing bool) *gorm.DB {
	// Initialize logger
	ceslogger.InitLogger("0", "test", "test", "test", "1.0")

	// Create in-memory database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto-migrate test table
	db.AutoMigrate(&model.WorkflowHistory{})

	if enableTracing {
		// Setup span exporter
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(
			trace.WithSyncer(exporter),
		)
		otel.SetTracerProvider(tp)

		// Initialize global tracer
		Tracer = tp.Tracer("benchmark")

		// Register GORM tracing
		RegisterGormTracing(db, "benchmark_db")
	}

	return db
}

// BenchmarkInsert_WithoutTracing measures INSERT performance without tracing
func BenchmarkInsert_WithoutTracing(b *testing.B) {
	db := setupBenchmarkDB(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := &model.WorkflowHistory{
			Id:                      fmt.Sprintf("id-%d", i),
			WorkflowConfigurationId: "workflow-id",
			Status:                  "EXECUTED",
		}
		db.Create(data)
	}
}

// BenchmarkInsert_WithTracing measures INSERT performance with tracing enabled
func BenchmarkInsert_WithTracing(b *testing.B) {
	db := setupBenchmarkDB(true)

	ctx := context.Background()
	_, span := Tracer.Start(ctx, "benchmark-parent")
	defer span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := &model.WorkflowHistory{
			Id:                      fmt.Sprintf("id-%d", i),
			WorkflowConfigurationId: "workflow-id",
			Status:                  "EXECUTED",
		}
		db.WithContext(ctx).Create(data)
	}
}

// BenchmarkSelect_WithoutTracing measures SELECT performance without tracing
func BenchmarkSelect_WithoutTracing(b *testing.B) {
	db := setupBenchmarkDB(false)

	// Seed data
	for i := 0; i < 100; i++ {
		db.Create(&model.WorkflowHistory{
			Id:     "test-id",
			Status: "EXECUTED",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []model.WorkflowHistory
		db.Find(&results)
	}
}

// BenchmarkSelect_WithTracing measures SELECT performance with tracing enabled
func BenchmarkSelect_WithTracing(b *testing.B) {
	db := setupBenchmarkDB(true)

	// Seed data
	for i := 0; i < 100; i++ {
		db.Create(&model.WorkflowHistory{
			Id:     "test-id",
			Status: "EXECUTED",
		})
	}

	ctx := context.Background()
	_, span := Tracer.Start(ctx, "benchmark-parent")
	defer span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []model.WorkflowHistory
		db.WithContext(ctx).Find(&results)
	}
}

// BenchmarkUpdate_WithoutTracing measures UPDATE performance without tracing
func BenchmarkUpdate_WithoutTracing(b *testing.B) {
	db := setupBenchmarkDB(false)

	// Seed data
	db.Create(&model.WorkflowHistory{
		Id:     "test-id",
		Status: "ACTIVE",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Model(&model.WorkflowHistory{}).
			Where("id = ?", "test-id").
			Update("status", "EXECUTED")
	}
}

// BenchmarkUpdate_WithTracing measures UPDATE performance with tracing enabled
func BenchmarkUpdate_WithTracing(b *testing.B) {
	db := setupBenchmarkDB(true)

	// Seed data
	db.Create(&model.WorkflowHistory{
		Id:     "test-id",
		Status: "ACTIVE",
	})

	ctx := context.Background()
	_, span := Tracer.Start(ctx, "benchmark-parent")
	defer span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.WithContext(ctx).Model(&model.WorkflowHistory{}).
			Where("id = ?", "test-id").
			Update("status", "EXECUTED")
	}
}

// BenchmarkDelete_WithoutTracing measures DELETE performance without tracing
func BenchmarkDelete_WithoutTracing(b *testing.B) {
	db := setupBenchmarkDB(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Seed data for each iteration
		id := fmt.Sprintf("del-%d", i)
		db.Create(&model.WorkflowHistory{
			Id:     id,
			Status: "FAILED",
		})

		// Delete
		db.Unscoped().Where("id = ?", id).Delete(&model.WorkflowHistory{})
	}
}

// BenchmarkDelete_WithTracing measures DELETE performance with tracing enabled
func BenchmarkDelete_WithTracing(b *testing.B) {
	db := setupBenchmarkDB(true)

	ctx := context.Background()
	_, span := Tracer.Start(ctx, "benchmark-parent")
	defer span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Seed data for each iteration
		id := fmt.Sprintf("del-%d", i)
		db.Create(&model.WorkflowHistory{
			Id:     id,
			Status: "FAILED",
		})

		// Delete
		db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&model.WorkflowHistory{})
	}
}

// BenchmarkComplexQuery_WithoutTracing measures complex query performance without tracing
func BenchmarkComplexQuery_WithoutTracing(b *testing.B) {
	db := setupBenchmarkDB(false)

	// Seed data
	for i := 0; i < 1000; i++ {
		db.Create(&model.WorkflowHistory{
			Id:     "test-id",
			Status: "EXECUTED",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []model.WorkflowHistory
		var totalData int64
		db.Model(&model.WorkflowHistory{}).
			Where("status = ?", "EXECUTED").
			Count(&totalData).
			Limit(10).
			Offset(0).
			Find(&results)
	}
}

// BenchmarkComplexQuery_WithTracing measures complex query performance with tracing enabled
func BenchmarkComplexQuery_WithTracing(b *testing.B) {
	db := setupBenchmarkDB(true)

	// Seed data
	for i := 0; i < 1000; i++ {
		db.Create(&model.WorkflowHistory{
			Id:     "test-id",
			Status: "EXECUTED",
		})
	}

	ctx := context.Background()
	_, span := Tracer.Start(ctx, "benchmark-parent")
	defer span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []model.WorkflowHistory
		var totalData int64
		db.WithContext(ctx).Model(&model.WorkflowHistory{}).
			Where("status = ?", "EXECUTED").
			Count(&totalData).
			Limit(10).
			Offset(0).
			Find(&results)
	}
}
