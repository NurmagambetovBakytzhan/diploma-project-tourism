package usecase

import (
	"github.com/google/uuid"
	"notification-service/internal/entity"
	"notification-service/internal/usecase/repo"
)

// TranslationUseCase -.
type NotificationsUseCase struct {
	repo *repo.NotificationRepo
}

// NewNotificationsUseCase -.
func NewNotificationsUseCase(r *repo.NotificationRepo) *NotificationsUseCase {
	return &NotificationsUseCase{
		repo: r,
	}
}

func (n *NotificationsUseCase) GetMyNotifications(userID uuid.UUID) ([]*entity.Notification, error) {
	return n.repo.GetMyNotifications(userID)
}
