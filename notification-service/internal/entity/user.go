package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	ID uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
}

type User struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"ID"`
	Username   string    `gorm:"unique;not null" json:"username"`
	Email      string    `gorm:"unique;not null" json:"email"`
	Password   string    `gorm:"not null" json:"password"`
	Role       string    `gorm:"not null" json:"role"` // user,admin, etc.
}

type Notification struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"ID"`
	UserID     string    `json:"userID" gorm:"type:uuid;default:uuid_generate_v4()" json:"userID"`
	ChatID     string    `json:"chatID" gorm:"type:uuid;default:uuid_generate_v4()"`
	Message    string    `json:"message" gorm:"not null" json:"message"`
	Topic      string    `json:"type" gorm:"not null" json:"type"`
}

type NotificationToWS struct {
	From    User   `json:"from"`
	To      User   `json:"to"`
	Message string `json:"message"`
	ChatID  string `json:"chat_id"`
}
