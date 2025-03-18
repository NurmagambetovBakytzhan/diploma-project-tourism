// Package usecase implements application business logic. Each logic group in own file.
package usecase

import "social-service/internal/entity"

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	SocialInterface interface {
		CreateChat(Chat *entity.CreateChatDTO) (*entity.Chat, error)
		EnterToChat(Chat *entity.EnterToChatDTO) (*entity.ChatParticipants, error)
	}
	KafkaMessageProcessor interface {
		ProcessMessage(key, value []byte) error
	}
)
