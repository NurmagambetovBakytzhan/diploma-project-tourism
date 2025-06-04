// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"authentication-service/internal/entity"
	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	UserInterface interface {
		LoginUser(user *entity.LoginUserDTO) (string, error)
		RegisterUser(user *entity.User) (*entity.User, string, error)
		GetUsers() ([]*entity.User, error)
		GetMe(id uuid.UUID) (*entity.User, error)
		VerifyEmail(sessionID, code string) error
	}
)
