package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tour struct {
	gorm.Model      `swaggerignore:"true"`
	ID              uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Description     string    `json:"description"`
	Route           string    `json:"route"`
	OwnerID         uuid.UUID `json:"owner_id" gorm:"type:uuid;index"`
	Name            string    `json:"name"`
	TelegramChatURL string    `json:"telegram_chat_url"`
	// Relationships
	TourImages        []Image         `json:"tour_images" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	TourVideos        []Video         `json:"tour_videos" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	TourPanoramas     []Panorama      `json:"tour_panoramas" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	TourEvents        []TourEvent     `json:"tour_events" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	TourCategories    []TourCategory  `json:"tour_categories" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	TourLocation      *TourLocation   `json:"tour_location" gorm:"foreignKey:TourID;references:ID"`
	TourUserFavorites []UserFavorites `json:"tour_user_favorites" gorm:"foreignKey:TourID;references:ID;constraint:OnDelete:CASCADE;"`
	AirpanoLink       string          `json:"airpano_link"`
}

type Category struct {
	gorm.Model     `swaggerignore:"true"`
	ID             uuid.UUID      `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name           string         `json:"name"`
	TourCategories []TourCategory `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TourCategory struct {
	CategoryID uuid.UUID `gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	Category   Category  `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TourID     uuid.UUID `json:"tour_id" gorm:"primaryKey;autoIncrement:false;type:uuid;index"`
	Tour       Tour      `gorm:"foreignKey:TourID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TourLocation struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`

	TourID uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	Tour   Tour      `gorm:"foreignKey:TourID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Image struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID     uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	Tour       Tour      `gorm:"foreignKey:TourID;constraint:OnDelete:CASCADE;"`
	ImageURL   string    `json:"image_url"`
}

type Video struct {
	gorm.Model `swaggerignore:"true"`
	ID         uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID     uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	Tour       Tour      `gorm:"foreignKey:TourID;constraint:OnDelete:CASCADE;"`
	VideoURL   string    `json:"video_url"`
}

type Panorama struct {
	gorm.Model  `swaggerignore:"true"`
	ID          uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID      uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	Tour        Tour      `gorm:"foreignKey:TourID;constraint:OnDelete:CASCADE;"`
	PanoramaURL string    `json:"panorama_url"`
}
