package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TourEvent struct {
	gorm.Model     `swaggerignore:"true"`
	ID             uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Tour           Tour
	Date           time.Time  `json:"data" gorm:"not null"`
	Price          float64    `json:"price" gorm:"not null"`
	Place          string     `json:"place" gorm:"not null"`
	AmountOfPlaces float64    `json:"amount" gorm:"not null"`
	IsOpened       bool       `json:"is_opened" gorm:"not null;default:true"`
	TourID         uuid.UUID  `json:"tour_id" gorm:"type:uuid;index"`
	Purchases      []Purchase `gorm:"foreignKey:TourEventID;references:ID"`
}
