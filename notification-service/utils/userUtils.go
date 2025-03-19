package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func GetUserIDFromContext(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil
	}

	// Convert user_id string to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return uuid.Nil
	}
	return userID
}
