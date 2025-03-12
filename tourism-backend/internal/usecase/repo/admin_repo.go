package repo

import (
	"fmt"
	"tourism-backend/internal/entity"
	"tourism-backend/pkg/postgres"
)

type AdminRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewAdminRepo(pg *postgres.Postgres) *AdminRepo {
	return &AdminRepo{pg}
}

func (r *AdminRepo) GetUsers() ([]*entity.User, error) {
	users := make([]*entity.User, 0)

	err := r.PG.Conn.Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}
