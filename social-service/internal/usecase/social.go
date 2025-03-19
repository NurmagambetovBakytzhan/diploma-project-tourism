package usecase

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"os"
	"social-service/internal/entity"
	"social-service/internal/usecase/repo"
)

type SocialUseCase struct {
	repo     *repo.SocialRepo
	producer sarama.SyncProducer
}

// NewTourismUseCase -.
// func NewSocialUseCase(r *repo.SocialRepo, p sarama.SyncProducer) *SocialUseCase {
func NewSocialUseCase(r *repo.SocialRepo) *SocialUseCase {
	return &SocialUseCase{
		repo: r,
		//producer: p,
	}
}
func (u *SocialUseCase) GetChatMessages(ChatID uuid.UUID) ([]*entity.Message, error) {
	return u.repo.GetChatMessages(ChatID)
}
func (u *SocialUseCase) PostMessage(message *entity.Message) {
	err := u.repo.PostMessage(message)
	if err != nil {
		log.Println(err)
		return
	}
	// Send Notification to Chat Participants

}

func (u *SocialUseCase) CheckChatParticipant(chatID uuid.UUID, userID uuid.UUID) bool {
	return u.repo.CheckChatParticipant(chatID, userID)
}

func (u *SocialUseCase) GetMyChats(userID uuid.UUID) ([]*entity.Chat, error) {
	return u.repo.GetMyChats(userID)
}

func (u *SocialUseCase) EnterToChat(Chat *entity.EnterToChatDTO) (*entity.ChatParticipants, error) {
	return u.repo.EnterToChat(Chat)
}

func (u *SocialUseCase) CreateChat(chat *entity.CreateChatDTO) (*entity.Chat, error) {
	return u.repo.CreateChat(chat)
}

func (u *SocialUseCase) PublishMessage(topic string, value interface{}) {
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
