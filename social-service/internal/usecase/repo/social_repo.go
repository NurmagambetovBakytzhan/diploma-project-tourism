package repo

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"social-service/internal/entity"
	"social-service/pkg/postgres"
)

type SocialRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewSocialRepo(pg *postgres.Postgres) *SocialRepo {
	return &SocialRepo{pg}
}

func (u *SocialRepo) EnterToChat(Chat *entity.EnterToChatDTO) (*entity.ChatParticipants, error) {
	chatToEnter := entity.ChatParticipants{
		ChatID: Chat.ChatID,
		UserID: Chat.UserID,
		Role:   "user",
	}
	err := u.PG.Conn.Create(&chatToEnter).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		log.Println("Error creating EnterToChat", err)
		return nil, errors.New("user already exists")
	}
	if err != nil {
		log.Println("Error creating EnterToChat: ", err)
		return nil, fmt.Errorf("error creating chat: %w", err)
	}
	return &chatToEnter, nil
}

func (u *SocialRepo) CreateChat(chat *entity.CreateChatDTO) (*entity.Chat, error) {
	chatToCreate := entity.Chat{
		Name:        chat.Name,
		Description: chat.Description,
		OwnerID:     chat.UserID,
	}
	err := u.PG.Conn.Transaction(func(tx *gorm.DB) error {
		err := u.PG.Conn.Create(&chatToCreate).Error
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return fmt.Errorf("user may not exists")
		}
		if err != nil {
			log.Println("Error creating chat: ", err)
			return fmt.Errorf("create chat: %w", err)
		}

		chatParticipant := entity.ChatParticipants{
			ChatID: chatToCreate.ID,
			UserID: chatToCreate.OwnerID,
			Role:   "admin",
		}
		err = u.PG.Conn.Create(&chatParticipant).Error
		if err != nil {
			log.Println("Error creating chat participant: ", err)
			return fmt.Errorf("create chat participant: %w", err)
		}
		return nil
	})
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return nil, fmt.Errorf("user may not exists")
	}
	if err != nil {
		log.Println("Error creating chat: ", err)
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return &chatToCreate, nil
}

func (u *SocialRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	err := u.PG.Conn.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
