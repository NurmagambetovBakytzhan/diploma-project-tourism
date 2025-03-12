package kafka

import (
	"github.com/IBM/sarama"
	"log"
)

func New(kafkaAddress string) (sarama.SyncProducer, error) {
	brokers := []string{kafkaAddress}
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
		return nil, err
	}
	return producer, nil
}
