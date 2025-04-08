// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"github.com/google/uuid"
	"social-service/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	SocialInterface interface {
		CreateChat(Chat *entity.CreateChatDTO) (*entity.Chat, error)
		EnterToChat(Chat *entity.EnterToChatDTO) (*entity.ChatParticipants, error)
		GetMyChats(userID uuid.UUID) ([]*entity.Chat, error)
		CheckChatParticipant(chatID uuid.UUID, userID uuid.UUID) bool
		PostMessage(message *entity.Message)
		GetChatMessages(chatID uuid.UUID) ([]*entity.Message, error)
		GetAllChats() ([]*entity.GetChatsDTO, error)
	}
	KafkaMessageProcessor interface {
		ProcessMessage(key, value []byte) error
	}
)
