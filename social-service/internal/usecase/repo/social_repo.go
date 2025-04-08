package repo

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
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

func (u *SocialRepo) GetAllChats() ([]*entity.GetChatsDTO, error) {
	var chats []entity.Chat
	var chatDTOs []*entity.GetChatsDTO

	err := u.PG.Conn.Select("id, name, description").Find(&chats).Error
	if err != nil {
		log.Println("GetAllChats: ", err)
		return nil, fmt.Errorf("Error getting all chats from DB: %v", err)
	}

	for _, chat := range chats {
		chatDTO := &entity.GetChatsDTO{
			ID:          chat.ID,
			Name:        chat.Name,
			Description: chat.Description,
		}
		chatDTOs = append(chatDTOs, chatDTO)
	}

	return chatDTOs, nil
}
func (u *SocialRepo) GetChatParticipants(chatID uuid.UUID) ([]uuid.UUID, error) {
	var results []uuid.UUID
	err := u.PG.Conn.Table("social_service.chat_participants").
		Select("user_id").
		Where("chat_id = ?", chatID).
		Find(&results).Error
	if err != nil {
		log.Println("Error GetChatParticipants: ", err)
		return nil, fmt.Errorf("error getting chat participants")
	}
	return results, nil
}

func (u *SocialRepo) GetChatMessages(ChatID uuid.UUID) ([]*entity.Message, error) {
	var chatMessages []*entity.Message

	err := u.PG.Conn.
		Where("chat_id = ?", ChatID).
		Order("created_at DESC").
		Find(&chatMessages).
		Error
	if err != nil {
		log.Println("Error getting chat Messages: ", err)
		return nil, err
	}
	return chatMessages, nil
}

func (u *SocialRepo) PostMessage(message *entity.Message) error {
	err := u.PG.Conn.Create(&message).Error
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (u *SocialRepo) CheckChatParticipant(chatID uuid.UUID, userID uuid.UUID) bool {
	var count int64
	err := u.PG.Conn.Table("social_service.chat_participants").
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	if err != nil {
		log.Println("Error CheckChatParticipant: ", err)
		return false
	}
	return count > 0
}

func (u *SocialRepo) GetMyChats(userID uuid.UUID) ([]*entity.Chat, error) {
	var chats []*entity.Chat

	err := u.PG.Conn.Joins("JOIN social_service.chat_participants ON social_service.chat_participants.chat_id = social_service.chats.id").
		Where("social_service.chat_participants.user_id = ?", userID).
		Find(&chats).Error
	if err != nil {
		log.Printf("GetMyChats Error: %v", err)
		return nil, fmt.Errorf("GetMyChats Error")
	}
	return chats, nil
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
