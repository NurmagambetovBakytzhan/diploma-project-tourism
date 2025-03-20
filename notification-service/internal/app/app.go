// Package app configures and runs application.
package app

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"notification-service/config"
	v1 "notification-service/internal/controller/http/v1"
	"notification-service/internal/usecase"
	"notification-service/internal/usecase/repo"
	"notification-service/pkg/kafka"
	"notification-service/pkg/logger"
	"notification-service/pkg/postgres"
	pkg "notification-service/pkg/websocket"
)

func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	go pkg.SocketHandler()

	// Repository
	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.NewNotificationPostGres: %w", err))
	}

	err = pg.Connect(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.connect: %w", err))
	}
	err = pg.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.Migrate: %w", err))
	}

	// kafka CONSUMER logic
	kafkaBrokers := []string{"kafka:9092"}
	groupID := "notification-group"
	topic := "notifications"

	processor := usecase.NewKafkaMessageProcessor(repo.NewNotificationRepo(pg))
	consumer, err := kafka.NewKafkaConsumer(kafkaBrokers, groupID, topic, processor)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.ConsumeMessages(ctx)

	// Use case
	notificationsUseCase := usecase.NewNotificationsUseCase(
		repo.NewNotificationRepo(pg),
	)

	// HTTP Server
	handler := fiber.New()

	// New Router
	v1.NewRouter(handler, l, notificationsUseCase)

	port := fmt.Sprintf(":%s", cfg.Port)
	// Waiting signal
	addr := flag.String("addr", port, "http service address")
	flag.Parse()
	handler.Listen(*addr)
}
