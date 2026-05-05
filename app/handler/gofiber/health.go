package gofiber

import (
	"context"
	"database/sql"
	"time"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soluixdeveloper/ces-orchestratorService/config/redis"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Services  map[string]string `json:"services"`
	Timestamp string            `json:"timestamp"`
}

func HealthCheck(c *fiber.Ctx) error {
	health := HealthStatus{
		Status: "healthy",
		Services: map[string]string{
			"database": checkDatabaseHealth(),
			"rabbitmq": checkRabbitMQHealth(),
			"redis":    checkRedisHealth(),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check if any service is unhealthy
	for _, status := range health.Services {
		if status != "healthy" {
			health.Status = "degraded"
			return c.Status(503).JSON(health)
		}
	}

	return c.JSON(health)
}

func ReadinessCheck(c *fiber.Ctx) error {
	// Simple readiness check - service is ready if it can respond
	return c.JSON(map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   config.AppConfig.ModuleVersion,
		"service":   config.AppConfig.ProjectName,
	})
}

func LivenessCheck(c *fiber.Ctx) error {
	// Liveness check - service is alive if it can respond
	return c.JSON(map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

var startTime = time.Now()

func checkDatabaseHealth() string {
	if config.PostgreDB == nil {
		return "unavailable"
	}

	// Cast to *sql.DB and use PingContext
	if sqlDB, ok := interface{}(config.PostgreDB).(*sql.DB); ok {
		if err := sqlDB.PingContext(context.Background()); err != nil {
			return "unhealthy"
		}
		return "healthy"
	}

	// If cast fails, assume unavailable
	return "unavailable"
}

func checkRabbitMQHealth() string {
	// Basic check - assume healthy if no errors during initialization
	// TODO: Implement proper connection check when cesrabbitmq supports it
	return "healthy"
}

func checkRedisHealth() string {
	if !config.AppConfig.RedisUse {
		return "disabled"
	}

	// Basic check - assume healthy if Redis is enabled
	// TODO: Implement proper Redis connection check
	return "healthy"
}

func livenessHandler(c *fiber.Ctx) error {
	fmt.Printf("[health/live] status=alive\n")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "alive",
	})
}

func checkRabbitMQ() error {
	conn, err := amqp.Dial(config.RabbitMQConfig.RabbitMQURL)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func startupHandler(c *fiber.Ctx) error {
	checks := fiber.Map{}
	ready := true

	if config.AppConfig.ProjectName == "" {
		checks["config"] = "not loaded"
		ready = false
	} else {
		checks["config"] = "ok"
	}

	if config.PostgreDB == nil {
		checks["db"] = "not initialized"
		ready = false
	} else {
		checks["db"] = "ok"
	}

	if !redis.IsInitialized() {
		checks["redis"] = "not initialized"
		ready = false
	} else {
		checks["redis"] = "ok"
	}

	if err := checkRabbitMQ(); err != nil {
		checks["rabbitmq"] = err.Error()
		ready = false
	} else {
		checks["rabbitmq"] = "ok"
	}

	status := "started"
	httpStatus := fiber.StatusOK
	if !ready {
		status = "starting"
		httpStatus = fiber.StatusServiceUnavailable
	}

	fmt.Printf("[health/startup] status=%s checks=%v\n", status, checks)

	return c.Status(httpStatus).JSON(fiber.Map{
		"status": status,
		"checks": checks,
	})
}

func readinessHandler(c *fiber.Ctx) error {
	deps := fiber.Map{}
	ready := true

	if config.PostgreDB == nil {
		deps["db"] = "not initialized"
		ready = false
	} else {
		sqlDB, err := config.PostgreDB.DB()
		if err != nil {
			deps["db"] = err.Error()
			ready = false
		} else if err = sqlDB.Ping(); err != nil {
			deps["db"] = err.Error()
			ready = false
		} else {
			deps["db"] = "ok"
		}
	}

	if err := redis.Ping(); err != nil {
		deps["redis"] = err.Error()
		ready = false
	} else {
		deps["redis"] = "ok"
	}

	if err := checkRabbitMQ(); err != nil {
		deps["rabbitmq"] = err.Error()
		ready = false
	} else {
		deps["rabbitmq"] = "ok"
	}

	status := "ready"
	httpStatus := fiber.StatusOK
	if !ready {
		status = "not ready"
		httpStatus = fiber.StatusServiceUnavailable
	}

	fmt.Printf("[health/ready] status=%s dependencies=%v\n", status, deps)

	return c.Status(httpStatus).JSON(fiber.Map{
		"status":       status,
		"dependencies": deps,
	})
}