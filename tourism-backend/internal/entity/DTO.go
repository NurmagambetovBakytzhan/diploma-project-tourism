package entity

import (
	"github.com/google/uuid"
	"time"
)

type WeatherInfo struct {
	TempCelsius    float64          `json:"temp_c"`
	TempFahrenheit float64          `json:"temp_f"`
	Condition      WeatherCondition `json:"condition"`
	WindKpH        float64          `json:"wind_kph"`
	WindMpH        float64          `json:"wind_mph"`
	WindDir        string           `json:"wind_dir"`
}

type WeatherInfoRQ struct {
	Date      string  `json:"date"`
	Time      string  `json:"time"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type WeatherCondition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
	Code int    `json:"code"`
}

type TourPurchaseRequest struct {
	TourEventID uuid.UUID `json:"tour_event_id"`
}

type CreateTourDTO struct {
	Description string `json:"description"`
	Route       string `json:"route"`
	Price       int    `json:"price"`
	//TourImages  []Image   `json:"tour_images"`
	//TourVideos  []Video   `json:"tour_videos"`
}

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

type CreateTourEventDTO struct {
	Date           time.Time `json:"date" gorm:"not null"`
	Price          float64   `json:"price" gorm:"not null"`
	Place          string    `json:"place" gorm:"not null"`
	TourID         uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	AmountOfPlaces float64   `json:"amount_of_places" gorm:"not null"`
}

type CreateTourCategoryDTO struct {
	TourID     uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	CategoryID uuid.UUID `json:"category_id" gorm:"type:uuid;index"`
}

type CreateTourLocationDTO struct {
	TourID    uuid.UUID `json:"tour_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

type TourEventFilter struct {
	CategoryIDs []uuid.UUID `json:"category_ids,omitempty"`
	StartDate   time.Time   `json:"start_date,omitempty"`
	EndDate     time.Time   `json:"end_date,omitempty"`
	MinPrice    float64     `json:"min_price,omitempty"`
	MaxPrice    float64     `json:"max_price,omitempty"`
}
