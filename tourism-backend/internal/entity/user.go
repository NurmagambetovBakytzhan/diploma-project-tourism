package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model          `swaggerignore:"true"`
	ID                  uuid.UUID  `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username            string     `gorm:"unique;not null"`
	Email               string     `gorm:"unique;not null"`
	Password            string     `gorm:"not null"`
	Role                string     `gorm:"not null"` // user,admin, etc.
	CreatedTours        []Tour     `gorm:"foreignKey:OwnerID;references:ID"`
	PurchasedTourEvents []Purchase `gorm:"foreignKey:UserID;references:ID"`
}
