package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model          `swaggerignore:"true"`
	ID                  uuid.UUID       `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username            string          `gorm:"unique;not null"`
	Email               string          `gorm:"unique;not null"`
	Password            string          `gorm:"not null"`
	Role                string          `gorm:"not null"` // user,admin, etc.
	CreatedTours        []Tour          `gorm:"foreignKey:OwnerID;references:ID"`
	PurchasedTourEvents []Purchase      `gorm:"foreignKey:UserID;references:ID"`
	FavoriteTours       []UserFavorites `gorm:"foreignKey:UserID;references:ID"`
}

type UserFavorites struct {
	UserID uuid.UUID `gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TourID uuid.UUID `json:"tour_id" gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	Tour   Tour      `gorm:"foreignKey:TourID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type UserActivity struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID     uuid.UUID `json:"user_id" gorm:"autoIncrement:false;type:uuid;index"`
	TourID     uuid.UUID `json:"tour_id" gorm:"autoIncrement:false;type:uuid;index"`
}
