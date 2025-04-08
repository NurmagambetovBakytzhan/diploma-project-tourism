package usecase

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"social-service/internal/entity"
	"social-service/internal/usecase/repo"
)

type SocialUseCase struct {
	repo     *repo.SocialRepo
	producer sarama.SyncProducer
}

// NewTourismUseCase -.
func NewSocialUseCase(r *repo.SocialRepo, p sarama.SyncProducer) *SocialUseCase {
	return &SocialUseCase{
		repo:     r,
		producer: p,
	}
}

func (u *SocialUseCase) GetAllChats() ([]*entity.GetChatsDTO, error) {
	return u.repo.GetAllChats()
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

	recipientList, err := u.repo.GetChatParticipants(message.ChatID)
	if err != nil {
		log.Println(err)
		return
	}
	//kafkaMessage := entity.NotificationMessageToKafkaDTO{
	//	Topic:      "MESSAGE",
	//	ChatID:     message.ChatID,
	//	AuthorID:   message.UserID,
	//	Message:    message.Text,
	//	Recipients: recipientList,
	//}
	kafkaMessage := entity.Notification{
		Topic: "MESSAGE",
		Data: map[string]interface{}{
			"ChatID":   message.ChatID,
			"AuthorID": message.UserID,
			"Message":  message.Text,
		},
		Recipients: recipientList,
	}
	// Send Notification to Chat Participants
	u.PublishMessage("notifications", kafkaMessage)

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
		log.Println("Failed to marshal message:", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonMessage),
	}

	_, _, err = u.producer.SendMessage(msg)
	if err != nil {
		log.Println("Failed to send message:", err)
	}
	log.Println("Message sent!")

}
