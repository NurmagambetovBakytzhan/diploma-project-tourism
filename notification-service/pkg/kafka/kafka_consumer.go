package kafka

import (
	"context"
	"log"
	"notification-service/internal/usecase"
	"time"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	processor     usecase.KafkaMessageProcessor
}

func NewKafkaConsumer(brokers []string, groupID, topic string, processor usecase.KafkaMessageProcessor) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0 // specify appropriate Kafka version
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		processor:     processor,
	}, nil
}

// Implement Sarama's ConsumerGroupHandler interface
func (kc *KafkaConsumer) ConsumeMessages(ctx context.Context) {
	handler := &consumerGroupHandler{processor: kc.processor}

	for {
		if err := kc.consumerGroup.Consume(ctx, []string{kc.topic}, handler); err != nil {
			log.Printf("Error consuming messages: %v", err)
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

type consumerGroupHandler struct {
	processor usecase.KafkaMessageProcessor
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		err := h.processor.ProcessMessage(message.Key, message.Value)
		if err == nil {
			session.MarkMessage(message, "")
		}
	}
	return nil
}
