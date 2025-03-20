// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/casbin/casbin/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"social-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	// Swagger docs.
	_ "social-service/docs"
	"social-service/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8060
// @BasePath    /v1
func NewRouter(handler *fiber.App, l logger.Interface, csbn *casbin.Enforcer, usecase *usecase.SocialUseCase) {
	// Options
	handler.Use(fiberLogger.New())
	//handler.Use(recover.New)

	// Swagger
	handler.Get("/v1/social/swagger/*", fiberSwagger.WrapHandler)

	// K8s probe
	handler.Get("/v1/social/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"Service": "Social Service!"})
	})

	// Prometheus metrics
	//handler.Get("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newSocialRoutes(h, l, csbn, usecase)
	}
}
