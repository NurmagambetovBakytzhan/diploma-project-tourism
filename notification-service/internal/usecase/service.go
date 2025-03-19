package usecase

type Service struct {
	NotificationUseCase *NotificationsUseCase
}

func NewService(tour *NotificationsUseCase) *Service {
	return &Service{
		NotificationUseCase: tour,
	}
}
