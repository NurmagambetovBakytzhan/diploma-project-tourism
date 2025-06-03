package rate_limit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type clientInfo struct {
	limiter  *rate.Limiter
	banUntil time.Time
}

var (
	mu        sync.Mutex
	clients   = make(map[string]*clientInfo)
	rateLimit = rate.Limit(1)
	burst     = 1000
	banTime   = 1 * time.Second // ban duration
)

func getClientLimiter(ip string) *clientInfo {
	mu.Lock()
	defer mu.Unlock()

	if client, exists := clients[ip]; exists {
		return client
	}

	clients[ip] = &clientInfo{
		limiter:  rate.NewLimiter(rateLimit, burst),
		banUntil: time.Time{},
	}
	return clients[ip]
}

func RateLimiter(c *gin.Context) {
	ip := c.ClientIP()
	client := getClientLimiter(ip)

	if time.Now().Before(client.banUntil) {
		c.JSON(http.StatusTooManyRequests,
			gin.H{"error": fmt.Sprintf("you are temporarily banned until %s",
				client.banUntil.Format(time.RFC3339))})
		c.Abort()
		return
	}

	if !client.limiter.Allow() {
		client.banUntil = time.Now().Add(banTime)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, you are now banned for a few minutes"})
		c.Abort()
		return
	}

	c.Next()
}
