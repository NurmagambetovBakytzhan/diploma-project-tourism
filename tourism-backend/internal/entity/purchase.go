package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model  `swaggerignore:"true"`
	ID          uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	User        User      `json:"User"`
	TourEvent   TourEvent `json:"TourEvent"`
	UserID      uuid.UUID `json:"UserID"`
	TourEventID uuid.UUID `json:"TourEventID"`
	Status      string    `json:"Status"`
}

type CreatePaymentIntentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type CreatePaymentIntentResponse struct {
	ClientSecret string `json:"clientSecret"`
}

type ConfirmPaymentRequest struct {
	PaymentIntentID string `json:"paymentIntentId"`
}
