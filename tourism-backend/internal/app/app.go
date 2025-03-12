// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tourism-backend/pkg/casbin"
	"tourism-backend/pkg/kafka"
	"tourism-backend/pkg/payment"

	"github.com/gin-gonic/gin"

	"tourism-backend/config"
	v1 "tourism-backend/internal/controller/http/v1"
	"tourism-backend/internal/usecase"
	"tourism-backend/internal/usecase/repo"
	"tourism-backend/pkg/httpserver"
	"tourism-backend/pkg/logger"
	"tourism-backend/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.NewTourismUseCase: %w", err))
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
	groupID := "consumer-group"
	topic := "users"

	processor := usecase.NewKafkaMessageProcessor(repo.NewUserRepo(pg))
	consumer, err := kafka.NewKafkaConsumer(kafkaBrokers, groupID, topic, processor)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.ConsumeMessages(ctx)

	// Use case
	tourismUseCase := usecase.NewTourismUseCase(
		repo.NewTourismRepo(pg),
	)
	adminUseCase := usecase.NewAdminUseCase(
		repo.NewAdminRepo(pg),
	)

	service := usecase.NewService(tourismUseCase, adminUseCase)

	// HTTP Server
	handler := gin.New()
	handler.Static("/v1/tours/uploads", "./uploads")
	handler.MaxMultipartMemory = 200 << 20

	// Casbin
	csbn := casbin.InitCasbin()

	// Payment Processor
	paymentProcessor := payment.NewPaymentProcessor(10, tourismUseCase)

	// New Router
	v1.NewRouter(handler, l, service, csbn, paymentProcessor)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

}
