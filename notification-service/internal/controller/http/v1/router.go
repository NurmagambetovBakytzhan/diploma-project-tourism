// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	// Swagger docs.
	_ "notification-service/docs"
	"notification-service/internal/usecase"
	"notification-service/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1/tours
func NewRouter(handler *fiber.App, l logger.Interface, usecase *usecase.NotificationsUseCase) {
	// Options
	handler.Use(fiberLogger.New())
	//handler.Use(recover.New)

	// Swagger
	handler.Get("/v1/notifications/swagger/*", fiberSwagger.WrapHandler)

	// K8s probe
	handler.Get("/v1/notifications/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"Service": "Notification Service!"})
	})

	// Prometheus metrics
	//handler.Get("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	// Routers
	h := handler.Group("/v1")
	{
		newNotificationRoutes(h, usecase, l)
	}

}
