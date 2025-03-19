package entity

import "github.com/google/uuid"

type NotificationMessageToKafkaDTO struct {
	ChatID     string      `json:"chatID" binding:"required"`
	AuthorID   string      `json:"authorID" binding:"required"`
	Recipients []uuid.UUID `json:"recipients" binding:"required"`
	Message    string      `json:"message" binding:"required"`
}
