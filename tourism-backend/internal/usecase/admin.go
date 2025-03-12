package usecase

import (
	"fmt"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase/repo"
)

// TranslationUseCase -.
type AdminUseCase struct {
	repo *repo.AdminRepo
}

// NewTourismUseCase -.
func NewAdminUseCase(r *repo.AdminRepo) *AdminUseCase {
	return &AdminUseCase{
		repo: r,
	}
}

func (a *AdminUseCase) GetUsers() ([]*entity.User, error) {
	users, err := a.repo.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}
