// Package app configures and runs application.
package app

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"social-service/internal/usecase"
	"social-service/internal/usecase/repo"
	"social-service/pkg/casbin"
	"social-service/pkg/kafka"
	"social-service/pkg/postgres"
	pkg "social-service/pkg/websocket"

	"social-service/config"
	v1 "social-service/internal/controller/http/v1"
	"social-service/pkg/logger"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	go pkg.SocketHandler()

	//Repository
	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres: %w", err))
	}

	err = pg.Connect(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.connect: %w", err))
	}

	err = pg.Migrate()
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.Migrate: %w", err))
	}

	// kafka Producer Logic
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.Kafka.Address)
	if err != nil {
		l.Fatal(fmt.Errorf("error creating kafka producer: %w", err))
	}

	// Use Case
	socialUseCase := usecase.NewSocialUseCase(
		repo.NewSocialRepo(pg),
		kafkaProducer,
	)

	// kafka CONSUMER logic
	kafkaBrokers := []string{"kafka:9092"}
	groupID := "social-consumer-group"
	topic := "users"

	processor := usecase.NewKafkaMessageProcessor(repo.NewSocialRepo(pg))
	consumer, err := kafka.NewKafkaConsumer(kafkaBrokers, groupID, topic, processor)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.ConsumeMessages(ctx)

	// HTTP Server
	handler := fiber.New()

	// Casbin
	csbn := casbin.InitCasbin()

	// NewKafkaProducer Router
	v1.NewRouter(handler, l, csbn, socialUseCase)

	port := fmt.Sprintf(":%s", cfg.Port)
	// Waiting signal
	addr := flag.String("addr", port, "http service address")
	flag.Parse()
	handler.Listen(*addr)

}
