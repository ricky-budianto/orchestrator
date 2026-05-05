package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/soluixdeveloper/ces-utilities/v2/cesapp"
	"github.com/soluixdeveloper/ces-utilities/v2/cesenv"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"

	"github.com/robfig/cron/v3"
	"github.com/soluixdeveloper/ces-orchestratorService/app/function"
	"github.com/soluixdeveloper/ces-orchestratorService/app/function/dal"
	"github.com/soluixdeveloper/ces-orchestratorService/app/handler/gofiber"
	rabbitmqservice "github.com/soluixdeveloper/ces-orchestratorService/app/handler/rabbitmq"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/elasticlog"
	"github.com/soluixdeveloper/ces-orchestratorService/config/redis"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"github.com/soluixdeveloper/ces-utilities/v2/cesrabbitmq"
)

func main() {
	// Load environment config first (but don't initialize DB yet)
	envPath := []string{"../../../.", "../.", "."}
	envModel := map[string]interface{}{
		"app-config":       &config.AppConfig,
		"telemetry-config": &config.TelemetryConfig,
	}
	cesenv.InitEnv("config", "env", envPath, envModel, false)

	// Initialize logger early
	ceslogger.InitLogger(
		config.AppConfig.LogLevel,
		config.AppConfig.Environment,
		config.AppConfig.ProjectName,
		config.AppConfig.ProjectModule,
		config.AppConfig.ModuleVersion)
	logging := ceslogger.Logger{}

	// Initialize telemetry BEFORE database connection for GORM tracing
	logging.LogInfo(fmt.Sprintf("is telemetry enabled %v", config.TelemetryConfig.TelemetryEnabled))
	if config.TelemetryConfig.TelemetryEnabled {
		telemetryConfig := telemetry.Config{
			ServiceName:    config.AppConfig.ProjectName,
			ServiceVersion: config.AppConfig.ModuleVersion,
			TempoEndpoint:  config.TelemetryConfig.TempoEndpoint,
			Environment:    config.AppConfig.Environment,
		}

		if err := telemetry.InitTelemetry(telemetryConfig); err != nil {
			logging.LogError("failed to initialize telemetry", err.Error())
		} else {
			if err := telemetry.CreateCustomMetrics(); err != nil {
				logging.LogError("failed to create custom metrics", err.Error())
			}
		}
	}

	// Now initialize all configurations and services (including DB with GORM tracing)
	config.InitConfig()

	// Jika RUN_MIGRATION di set dan value nya true, start migration-only job —> return
	// Jika RUN_MIGRATION tidak di set (default), migration tetap berjalan tapi service lain tetap dijalankan
	if _, isSet := os.LookupEnv("RUN_MIGRATION"); isSet && config.AppConfig.RunMigration {
		ceslogger.NewLogger("").LogInfo("RunMigration: migration job finished, others service will not start")
		return
	}

	elasticlog.ElasticLog = elasticlog.NewLogger()
	redis.InitRedisClient()
	function.WorkflowInit()

	// Initialize RabbitMQ
	cesrabbitmq.InitRabbitMQ(
		config.AppConfig.ProjectName,
		config.AppConfig.ProjectModule,
		config.RabbitMQConfig.RabbitMQURL,
		30*time.Second,
		0,
		0,
	)
	go cesrabbitmq.ConsumeRPC(rabbitmqservice.RPCHandler)

	defer func() {
		if config.PostgreDB != nil {
			if sqlDB, err := config.PostgreDB.DB(); err == nil {
				sqlDB.Close()
			}
		}
	}()

	shutdownWG := sync.WaitGroup{}
	shutdownWG.Add(1)

	cesapp.OnShutdown().
		WillExecute(
			func(ctx context.Context, signal os.Signal) error {
				return cesapp.SetAppStatus(cesapp.AppStatusShuttingDown)
			},
		).
		AndThenExecute(
			func(ctx context.Context, signal os.Signal) error {
				return gofiber.ShutdownGofiber(ctx)
			},
		).
		AndThenExecute(
			func(ctx context.Context, signal os.Signal) error {
				return cesrabbitmq.GracefulShutdown()
			},
		).
		WithDebugLog().
		StartAsync(&shutdownWG)

	cesapp.SetAppStatus(cesapp.AppStatusReady)
	gofiber.InitGofiber() // blocking

	shutdownWG.Wait()
}

func consumeWithContext(ctx context.Context, logging *ceslogger.Logger) {
	for {
		select {
		case <-ctx.Done():
			logging.LogInfo("RabbitMQ consumer stopping...")
			return
		default:
			// Run the RPC consumer
			cesrabbitmq.ConsumeRPC(rabbitmqservice.RPCHandler)
			// If ConsumeRPC exits, wait a bit before retrying
			time.Sleep(time.Second)
		}
	}
}

func cleanlogWithContext(ctx context.Context, shutdownChan chan struct{}, logging *ceslogger.Logger) {
	c := cron.New()

	// Run at 2:00 AM every day
	_, err := c.AddFunc("0 2 * * *", func() {
		workflowHistories := []string{}
		err := dal.ListCleanUpWorkflowHistory(ctx, &workflowHistories)
		if err != nil {
			logging.LogError("error get list workflow histories", err.Error())
		}
		err = dal.DeleteWorkflowHistory(ctx, workflowHistories)
		if err != nil {
			logging.LogError("error delete workflow histories", err.Error())
		}
		workflowStates := []string{}
		err = dal.ListCleanUpWorkflowState(ctx, &workflowStates)
		if err != nil {
			logging.LogError("error get list workflow state", err.Error())
		}
		err = dal.DeleteWorkflowState(ctx, workflowStates)
		if err != nil {
			logging.LogError("error delete workflow state", err.Error())
		}
		logging.LogInfo("CRON COMPLETED ", len(workflowHistories)+len(workflowStates))
	})
	if err != nil {
		logging.LogError("Failed to schedule job:", err.Error())
		return
	}

	// Start the cron scheduler
	c.Start()

	// Wait for shutdown signal or context cancellation
	select {
	case <-shutdownChan:
		logging.LogInfo("Cron scheduler stopping...")
	case <-ctx.Done():
		logging.LogInfo("Cron scheduler stopping due to context cancellation...")
	}

	// Stop the cron scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Stop()
	logging.LogInfo("Cron scheduler stopped")
}

func cleanupResources(ctx context.Context, logging *ceslogger.Logger) {
	// Cleanup telemetry
	if config.TelemetryConfig.TelemetryEnabled && telemetry.TelemetryCleanup != nil {
		if err := telemetry.TelemetryCleanup(ctx); err != nil {
			logging.LogError("error shutting down telemetry", err.Error())
		} else {
			logging.LogInfo("Telemetry shutdown completed")
		}
	}

	// Cleanup database connections
	if config.PostgreDB != nil {
		if sqlDB, err := config.PostgreDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logging.LogError("error closing database connection", err.Error())
			} else {
				logging.LogInfo("Database connection closed")
			}
		}
	}

	// Note: RabbitMQ and Redis connections are managed by their respective libraries
	// and should be cleaned up automatically when the process exits
}
