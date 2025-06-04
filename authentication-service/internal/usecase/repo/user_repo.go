package repo

import (
	"authentication-service/internal/entity"
	"authentication-service/pkg/postgres"
	"fmt"
	"github.com/google/uuid"
	"log"
)

type UserRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (u *UserRepo) GetMe(id uuid.UUID) (*entity.User, error) {
	var user entity.User

	err := u.PG.Conn.Find(&user, "id = ?", id).Error
	if err != nil {
		log.Println("Error GetMe: ", err)
		return nil, fmt.Errorf("Error Getting Your Information")
	}
	return &user, nil
}

func (u *UserRepo) GetUsers() ([]*entity.User, error) {
	users := make([]*entity.User, 0)

	err := u.PG.Conn.Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
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

func (u *UserRepo) VerifyUser(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	result := u.PG.Conn.Model(&entity.User{}).Where("id = ?", id).Update("verified", true)
	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch the updated user
	err := u.PG.Conn.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
