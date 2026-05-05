package gofiber

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/soluixdeveloper/ces-orchestratorService/app/handler/gofiber/apiv1"
	"github.com/soluixdeveloper/ces-orchestratorService/app/handler/gofiber/apiv2"
	"github.com/soluixdeveloper/ces-orchestratorService/app/middleware"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var app *fiber.App

func ShutdownGofiber(ctx context.Context) error {
	if app == nil {
		return nil
	}
	return app.ShutdownWithContext(ctx)
}

func InitGofiber() {

	app = fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           60 * time.Second,
	})

	// Or extend your config for customization
	if config.AppConfig.BackendCORS != "" {
		app.Use(cors.New(cors.Config{
			Next:             nil,
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
			AllowHeaders:     "",
			AllowCredentials: false,
			ExposeHeaders:    "",
			MaxAge:           0,
		}))
	}
	setupRoutes(app)

	if err := app.Listen(fmt.Sprintf(":%v", config.AppConfig.AppPort)); err != nil {
		log.Fatal("gofiber listen: ", err)
	}
}

func printInfo(c *fiber.Ctx) error {
	// return metadata project
	return c.SendString(fmt.Sprintf("== %s v%s ==", utils.AppName, utils.AppVersion))
}

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return printInfo(c)
	})

	// Health check endpoints for Kubernetes
	// app.Get("/health", HealthCheck)
	// app.Get("/ready", ReadinessCheck)
	// app.Get("/live", LivenessCheck)

	// Add metrics endpoint if telemetry is enabled
	if config.TelemetryConfig.TelemetryEnabled && telemetry.MetricsHandler != nil {
		app.Get("/metrics", adaptor.HTTPHandler(telemetry.MetricsHandler))
	}

	// Add telemetry middleware globally if enabled
	if config.TelemetryConfig.TelemetryEnabled {
		app.Use(middleware.TelemetryMiddleware())
	}

	health := app.Group("/health")

	health.Get("/live", livenessHandler)
	health.Get("/ready", readinessHandler)
	health.Get("/startup", startupHandler)

	api := app.Group(utils.PathService)
	api.Use(logger.New())

	// Add your other routes here
	// API v1
	apiv1.ApiV1Orchestration(app)
	apiv1.ApiV1(api)
	apiv2.ApiV2(api)
}
