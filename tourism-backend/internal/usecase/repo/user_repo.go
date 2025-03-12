package repo

import (
	"fmt"
	"tourism-backend/internal/entity"
	"tourism-backend/pkg/postgres"
)

type UserRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (u *UserRepo) LoginUser(user *entity.LoginUserDTO) (*entity.User, error) {

	var userFromDB entity.User

	if err := u.PG.Conn.Where("username = ?", user.Username).First(&userFromDB).Error; err != nil {
		return nil, fmt.Errorf("Username Not Found: %v", err)
	}
	return &userFromDB, nil
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	err := u.PG.Conn.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
