package usecase

import (
	"encoding/json"
	"log"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase/repo"
)

type kafkaMessageProcessor struct {
	repo *repo.UserRepo
}

func NewKafkaMessageProcessor(r *repo.UserRepo) KafkaMessageProcessor {
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
