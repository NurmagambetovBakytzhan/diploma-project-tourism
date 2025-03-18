package usecase

import (
	"encoding/json"
	"log"
	"social-service/internal/entity"
	"social-service/internal/usecase/repo"
)

type kafkaMessageProcessor struct {
	repo *repo.SocialRepo
}

func NewKafkaMessageProcessor(r *repo.SocialRepo) KafkaMessageProcessor {
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
	user := &entity.User{}
	err := json.Unmarshal(value, user)
	if err != nil {
		log.Printf("Error unmarshalling user: %v", err)
	}
	_, err = p.repo.RegisterUser(user)
	if err != nil {
		log.Printf("Error registering user: %v", err)
	}
}
