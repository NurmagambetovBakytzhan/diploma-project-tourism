// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/casbin/casbin/v2"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "authentication-service/docs"
	"authentication-service/internal/usecase"
	"authentication-service/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @Security BearerAuth
func NewRouter(handler *gin.Engine, l logger.Interface, usecase *usecase.UserUseCase, csbn *casbin.Enforcer) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	handler.GET("/v1/users/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/v1/users/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Service": "Auth Service!"})
	})

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newUserRoutes(h, usecase, l, csbn)
	}
}
