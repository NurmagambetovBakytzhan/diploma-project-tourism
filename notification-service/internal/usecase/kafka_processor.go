package usecase

import (
	"encoding/json"
	"log"
	"notification-service/internal/entity"
	"notification-service/internal/usecase/repo"
	pkg "notification-service/pkg/websocket"
)

type kafkaMessageProcessor struct {
	repo *repo.NotificationRepo
}

func NewKafkaMessageProcessor(r *repo.NotificationRepo) KafkaMessageProcessor {
	return &kafkaMessageProcessor{
		repo: r,
	}
}

func (p *kafkaMessageProcessor) ProcessMessage(key, value []byte) error {
	log.Printf("Processing Kafka message: key=%s, value=%s", key, value)
	go p.ProcessUser(value)
	return nil
}

func (p *kafkaMessageProcessor) ProcessUser(value []byte) {
	notificationsEvent := &entity.NotificationMessageToKafkaDTO{}
	err := json.Unmarshal(value, notificationsEvent)
	if err != nil {
		log.Printf("Error unmarshalling user: %v", err)
	}

	for _, recipient := range notificationsEvent.Recipients {
		pkg.Broadcast <- pkg.BroadcastObject{
			MSG: notificationsEvent.Message,
			FROM: pkg.ClientObject{
				ChatID: notificationsEvent.ChatID,
				UserID: notificationsEvent.AuthorID,
			},
			ChatId:    notificationsEvent.ChatID,
			Recipient: recipient.String(),
		}

		notification := entity.Notification{
			UserID:  notificationsEvent.AuthorID,
			ChatID:  notificationsEvent.ChatID,
			Message: notificationsEvent.Message,
		}
		err := p.repo.CreateNotification(&notification)
		if err != nil {
			log.Printf("Error creating Kafka message: %v, %v", err, notification)
		}
	}

}
