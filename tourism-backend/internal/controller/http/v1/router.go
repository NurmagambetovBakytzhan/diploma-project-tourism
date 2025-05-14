// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/stripe/stripe-go/v82"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"tourism-backend/config"
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
func NewRouter(handler *gin.Engine, l logger.Interface, service *usecase.Service, csbn *casbin.Enforcer, paymentProcessor *payment.PaymentProcessor, cfg *config.Config) {
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
	stripe.Key = cfg.Stripe.SecretKey
	if stripe.Key == "" {
		fmt.Println("Please set STRIPE_SECRET_KEY environment variable")
		return
	}
	fmt.Println("Stripe secret key: ", stripe.Key)
	// Routers
	h := handler.Group("/v1")
	{
		newTourismRoutes(h, service.TourUseCase, l, csbn, paymentProcessor)
		newAdminRoutes(h, service.AdminUseCase, l, csbn)
	}

}
