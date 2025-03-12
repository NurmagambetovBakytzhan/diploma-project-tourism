package usecase

type Service struct {
	TourUseCase  *TourismUseCase
	AdminUseCase *AdminUseCase
}

func NewService(tour *TourismUseCase, admin *AdminUseCase) *Service {
	return &Service{
		TourUseCase:  tour,
		AdminUseCase: admin,
	}
}
