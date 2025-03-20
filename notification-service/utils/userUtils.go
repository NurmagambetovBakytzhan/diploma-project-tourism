package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
)

func GetUserIDFromContext(c *fiber.Ctx) uuid.UUID {
	userIDStr := c.Locals("userID")

	// Convert user_id string to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		err := c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
		if err != nil {
			log.Println(err)
			return [16]byte{}
		}
		return uuid.Nil
	}
	return userID
}
