// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/casbin/casbin/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"tourism-backend/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// Swagger docs.
	_ "tourism-backend/docs"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1/tours
func NewRouter(handler *gin.Engine, l logger.Interface, service *usecase.Service, csbn *casbin.Enforcer, paymentProcessor *payment.PaymentProcessor) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	handler.GET("/v1/tours/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/v1/tours/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Service": "Tourism Service!"})
	})
	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newTourismRoutes(h, service.TourUseCase, l, csbn, paymentProcessor)
		newAdminRoutes(h, service.AdminUseCase, l, csbn)
	}

}
