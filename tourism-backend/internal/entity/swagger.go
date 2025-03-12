package entity

import (
	"github.com/google/uuid"
	"time"
)

type TourDocs struct {
	ID          uuid.UUID   `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" gorm:"index"`
	Description string      `json:"description"`
	Route       string      `json:"route"`
	Price       int         `json:"price"`
	TourImages  []ImageDocs `json:"tour_images" gorm:"foreignKey:TourID;references:ID"`
	TourVideos  []VideoDocs `json:"tour_videos" gorm:"foreignKey:TourID;references:ID"`
}

type ImageDocs struct {
	ID       uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID   uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	ImageURL string    `json:"image_bytes"`
}

type VideoDocs struct {
	ID       uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID   uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	VideoURL string    `json:"video_bytes"`
}
