package entity

import "github.com/google/uuid"

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional: "user" (default) or "admin"
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateChatDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	UserID      uuid.UUID
}

type ChatIdStringDTO struct {
	ChatID string `json:"chat_id"`
}

type EnterToChatDTO struct {
	ChatID uuid.UUID `json:"chatID" binding:"required"`
	UserID uuid.UUID
}

type NotificationMessageToKafkaDTO struct {
	Topic      string      `json:"topic" binding:"required"`
	ChatID     uuid.UUID   `json:"chatID" binding:"required"`
	AuthorID   uuid.UUID   `json:"authorID" binding:"required"`
	Recipients []uuid.UUID `json:"recipients" binding:"required"`
	Message    string      `json:"message" binding:"required"`
}
type Notification struct {
	Topic      string                 `json:"topic" binding:"required"`
	Data       map[string]interface{} `json:"data" binding:"required"`
	Recipients []uuid.UUID            `json:"recipients" binding:"required"`
}
