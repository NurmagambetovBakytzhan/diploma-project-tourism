// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"github.com/google/uuid"
	"notification-service/internal/entity"
)

type (
	KafkaMessageProcessor interface {
		ProcessMessage(key, value []byte) error
	}
	NotificationInterface interface {
		GetMyNotifications(userID uuid.UUID) ([]*entity.Notification, error)
	}
)
