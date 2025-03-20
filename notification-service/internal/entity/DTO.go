package entity

import "github.com/google/uuid"

type NotificationMessageToKafkaDTO struct {
	Topic      string      `json:"topic" binding:"required"`
	ChatID     string      `json:"chatID" binding:"required"`
	AuthorID   string      `json:"authorID" binding:"required"`
	Recipients []uuid.UUID `json:"recipients" binding:"required"`
	Message    string      `json:"message" binding:"required"`
}

type NotificationDTO struct {
	Topic      string                 `json:"topic" binding:"required"`
	Data       map[string]interface{} `json:"data" binding:"required"`
	Recipients []uuid.UUID            `json:"recipients" binding:"required"`
}
