package usecase

import (
	"encoding/json"
	"github.com/google/uuid"
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
	go p.ProcessNotification(value)
	return nil
}

func (p *kafkaMessageProcessor) ProcessNotification(value []byte) {
	notification := &entity.NotificationDTO{}
	err := json.Unmarshal(value, notification)
	if err != nil {
		log.Printf("Error unmarshalling user: %v", err)
	}

	for _, recipient := range notification.Recipients {
		pkg.Broadcast <- pkg.BroadcastObject{
			MSG:       notification.Data,
			Recipient: recipient.String(),
		}
		switch topic := notification.Topic; topic {
		case "MESSAGE":
			notification := entity.Notification{
				UserID:      notification.Data["AuthorID"].(string),
				ChatID:      notification.Data["ChatID"].(string),
				Message:     notification.Data["Message"].(string),
				Topic:       notification.Topic,
				RecipientID: recipient,
			}
			err := p.repo.CreateNotification(&notification)
			if err != nil {
				log.Printf("Error creating Kafka message: %v, %v", err, notification)
			}
		case "PAYMENT":
			notification := entity.Notification{
				UserID:      uuid.Nil.String(),
				ChatID:      uuid.Nil.String(),
				Topic:       notification.Topic,
				RecipientID: recipient,
			}
			err := p.repo.CreateNotification(&notification)
			if err != nil {
				log.Printf("Error creating Kafka message: %v, %v", err, notification)
			}
		default:
			log.Printf("Error processing Kafka message: %v, %v", err, notification)
		}

	}

}
