package usecase

import (
	"authentication-service/internal/entity"
	"authentication-service/internal/usecase/repo"
	"authentication-service/utils"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"os"
)

type UserUseCase struct {
	repo     *repo.UserRepo
	producer sarama.SyncProducer
}

// NewTourismUseCase -.
func NewUserUseCase(r *repo.UserRepo, p sarama.SyncProducer) *UserUseCase {
	return &UserUseCase{
		repo:     r,
		producer: p,
	}
}

func (a *UserUseCase) GetUsers() ([]*entity.User, error) {
	users, err := a.repo.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}

func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
	userFromRepo, err := u.repo.LoginUser(user)
	if err != nil {
		return "", fmt.Errorf("User From Repo: %w", err)
	}
	if !utils.CheckPassword(userFromRepo.Password, user.Password) {
		return "", fmt.Errorf("Check Password: %w", err)
	}
	token, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return "", fmt.Errorf("Generate JWT: %w", err)
	}
	return token, nil
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, error) {
	user, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	go u.PublishMessage("users", user)
	return user, nil
}

func (u *UserUseCase) PublishMessage(topic string, value interface{}) {
	jsonMessage, err := json.Marshal(value)
	if err != nil {
		log.Fatalln("Failed to marshal message:", err)
		os.Exit(1)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonMessage),
	}

	_, _, err = u.producer.SendMessage(msg)
	if err != nil {
		log.Fatalln("Failed to send message:", err)
		os.Exit(1)
	}
	log.Println("Message sent!")

}
