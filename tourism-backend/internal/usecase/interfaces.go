// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"github.com/google/uuid"
	"mime/multipart"
	"tourism-backend/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Tourism -.
	TourismInterface interface {
		// CreateTour GetTourByID(ctx context.Context, id uuid.UUID) (entity.Tour, error)
		CreateTour(tour *entity.Tour, imageFiles []*multipart.FileHeader, videFiles []*multipart.FileHeader) (*entity.Tour, error)
		GetTours() ([]entity.Tour, error)
		GetTourByID(ID string) (*entity.Tour, error)
		GetAllCategories() ([]entity.Category, error)
		CreateTourEvent(tourEvent *entity.TourEvent) (*entity.TourEvent, error)
		CheckTourOwner(tourID uuid.UUID, userID uuid.UUID) bool
		PayTourEvent(purchase *entity.Purchase) error
		CreatePurchase(purchase *entity.Purchase) (*entity.Purchase, error)
		CreateTourCategory(tourCategory *entity.CreateTourCategoryDTO) (*entity.TourCategory, error)
		CreateTourLocation(tourLocation *entity.CreateTourLocationDTO) (*entity.TourLocation, error)
		GetTourLocationByID(id uuid.UUID) (*entity.TourLocation, error)
		GetFilteredTourEvents(*entity.TourEventFilter) ([]*entity.TourEvent, error)
		GetWeatherByTourEventID(tourEventID uuid.UUID) (*entity.WeatherInfo, error)
		GetTourEventByID(id uuid.UUID) (*entity.TourEvent, error)
		AddFilesToTourByTourID(panoramasEntity []*entity.Panorama) ([]*entity.Panorama, error)
	}
	UserInterface interface {
		LoginUser(user *entity.LoginUserDTO) (string, error)
		RegisterUser(user *entity.User) (*entity.User, error)
	}
	AdminInterface interface {
		GetUsers() ([]*entity.User, error)
	}
	KafkaMessageProcessor interface {
		ProcessMessage(key, value []byte) error
	}
)
