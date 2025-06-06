package v1

import (
	"api-gateway/pkg/logger"
	rate_limit "api-gateway/pkg/rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type TourismRoutes struct {
	l logger.Interface
}

// ReverseProxy forwards requests to the target service
func ReverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Build target URL
		targetURL, err := url.Parse(target + c.Request.RequestURI)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
			return
		}

		// Create a new request with the original method
		req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers from original request
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Set(key, value)
			}
		}

		if isImageRequest(c.Request.URL.Path) {
			c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
			c.Header("Cache-Control", "public, max-age=31536000")
		}

		// Make the request to the target microservice
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach service"})
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		origin := c.Request.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:4200" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		// Set status code and write response
		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}

func isImageRequest(path string) bool {
	return strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".gif")
}
func NewRoutes(router *gin.Engine, l logger.Interface) {
	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:4200"},
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000" || origin == "http://localhost:4200"
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}), rate_limit.RateLimiter)

	// Load service URLs from environment variables
	//tourismAPI := os.Getenv("TOURISM_API_PORT") // Example: http://localhost:8080
	//authAPI := os.Getenv("AUTH_API_PORT")       // Example: http://localhost:8090
	tourismAPI := "http://tourism-backend:8080" // Example: http://localhost:8080
	authAPI := "http://auth-service:8090"
	socialAPI := "http://social-service:8060"
	notificationsAPI := "http://notification-service:8070"
	recommendationsAPI := "http://recommendation-service:8050"

	// Auth Service Routes
	router.Any("/v1/users/*any", ReverseProxy(authAPI))
	router.Any("/v1/admin/*any", ReverseProxy(tourismAPI))
	router.Any("/v1/tours/*any", ReverseProxy(tourismAPI)) // Forward to Tourism Service
	router.Any("/v1/social/*any", ReverseProxy(socialAPI))
	router.Any("/v1/notifications/*any", ReverseProxy(notificationsAPI))
	router.Any("/v1/recommendations/*any", ReverseProxy(recommendationsAPI))
}
