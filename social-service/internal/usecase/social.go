package usecase

import (
	"encoding/json"
	"github.com/IBM/sarama"
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
