package usecase

import (
	"authentication-service/internal/entity"
	"authentication-service/internal/usecase/repo"
	"authentication-service/pkg/email"
	"authentication-service/pkg/redis"
	"authentication-service/utils"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"os"
	"time"
)

type UserUseCase struct {
	repo     *repo.UserRepo
	producer sarama.SyncProducer
	redis    *redis.RedisClient
	email    *email.EmailService
}

// NewTourismUseCase -.
func NewUserUseCase(r *repo.UserRepo, p sarama.SyncProducer, redis *redis.RedisClient, email *email.EmailService) *UserUseCase {
	return &UserUseCase{
		repo:     r,
		producer: p,
		redis:    redis,
		email:    email,
	}
}

func (a *UserUseCase) GetMe(id uuid.UUID) (*entity.User, error) {
	return a.repo.GetMe(id)
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

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	user, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}

	// Generate session ID and verification code
	sessionID := uuid.New().String()
	code, err := email.GenerateVerificationCode()
	if err != nil {
		return nil, "", fmt.Errorf("generate verification code: %w", err)
	}

	// Store in Redis (expires in 15 minutes)
	verificationData := map[string]interface{}{
		"email":     user.Email,
		"code":      code,
		"user_id":   user.ID.String(),
		"expiresAt": time.Now().Add(15 * time.Minute).Unix(),
	}
	jsonData, err := json.Marshal(verificationData)
	if err != nil {
		return nil, "", fmt.Errorf("marshal verification data: %w", err)
	}
	err = u.redis.Set(sessionID, jsonData, 15*time.Minute)
	if err != nil {
		return nil, "", fmt.Errorf("store verification data: %w", err)
	}

	// Send email
	err = u.email.SendVerificationCode(user.Email, code)
	if err != nil {
		return nil, "", fmt.Errorf("send verification email: %w", err)
	}

	return user, sessionID, nil
}

func (u *UserUseCase) VerifyEmail(sessionID, code string) error {
	data, err := u.redis.Get(sessionID)
	if err != nil {
		return fmt.Errorf("invalid or expired session")
	}

	var verificationData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &verificationData); err != nil {
		return fmt.Errorf("invalid verification data")
	}

	if verificationData["code"] != code {
		return fmt.Errorf("invalid verification code")
	}

	// Mark user as verified in database
	userID, err := uuid.Parse(verificationData["user_id"].(string))
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	// Update user verification status in your repo
	// You'll need to add this method to your UserRepo
	verifiedUser, err := u.repo.VerifyUser(userID)
	if err != nil {
		return fmt.Errorf("verify user: %w", err)
	}

	go u.PublishMessage("users", verifiedUser)
	// Clean up
	_ = u.redis.Delete(sessionID)
	//go u.PublishMessage("users", user)

	return nil
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
