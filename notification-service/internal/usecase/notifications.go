package usecase

import (
	"notification-service/internal/usecase/repo"
)

// TranslationUseCase -.
type NotificationsUseCase struct {
	repo *repo.NotificationsRepo
}

// NewNotificationsUseCase -.
func NewNotificationsUseCase(r *repo.NotificationsRepo) *NotificationsUseCase {
	return &NotificationsUseCase{
		repo: r,
	}
}
