// Package usecase implements application business logic. Each logic group in own file.
package usecase

type (
	KafkaMessageProcessor interface {
		ProcessMessage(key, value []byte) error
	}
)
