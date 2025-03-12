package v1

import (
	"api-gateway/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
)

type TourismRoutes struct {
	l logger.Interface
}

// ReverseProxy forwards requests to the target service
func ReverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Build target URL
		targetURL, err := url.Parse(target + c.Request.RequestURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
			return
		}

		// Create a new request with the original method
		req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers from original request
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make the request to the target microservice
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach service"})
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// Set status code and write response
		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
func NewRoutes(router *gin.Engine, l logger.Interface) {
	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Load service URLs from environment variables
	//tourismAPI := os.Getenv("TOURISM_API_PORT") // Example: http://localhost:8080
	//authAPI := os.Getenv("AUTH_API_PORT")       // Example: http://localhost:8090
	tourismAPI := "http://tourism-backend:8080" // Example: http://localhost:8080
	authAPI := "http://auth-service:8090"

	if tourismAPI == "" || authAPI == "" {
		panic("Environment variables for microservices (TOURISM_API_URL, AUTH_API_URL, PAYMENT_API_URL) must be set")
	}

	// Auth Service Routes
	router.Any("/v1/users/*any", ReverseProxy(authAPI))
	router.Any("/v1/admin/*any", ReverseProxy(tourismAPI))
	router.Any("/v1/tours/*any", ReverseProxy(tourismAPI)) // Forward to Tourism Service

}
