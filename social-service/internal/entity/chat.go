package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model  `swaggerignore:"true"`
	ID          uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`

	ChatMessages []Message `json:"ChatMessages" gorm:"foreignKey:ChatID;references:ID;constraint:OnDelete:CASCADE;"`
	OwnerID      uuid.UUID `json:"owner_id" gorm:"type:uuid;index"`
	Owner        User      `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type ChatParticipants struct {
	ChatID uuid.UUID `json:"chat_id" gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	Chat   Chat      `gorm:"foreignKey:ChatID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID uuid.UUID `json:"user_id" gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE;"`
	Role   string    `json:"role" gorm:"not null;default:'user'"`
}

type Message struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Text       string    `json:"text"`
	UserID     uuid.UUID `json:"UserID" gorm:"type:uuid"`
	ChatID     uuid.UUID `json:"ChatID" gorm:"type:uuid;index"`
	Chat       Chat      `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;"`
	User       User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}
