// Package app configures and runs application.
package app

import (
	"authentication-service/pkg/casbin"
	"authentication-service/pkg/kafka"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"authentication-service/config"
	v1 "authentication-service/internal/controller/http/v1"
	"authentication-service/internal/usecase"
	"authentication-service/internal/usecase/repo"
	"authentication-service/pkg/httpserver"
	"authentication-service/pkg/logger"
	"authentication-service/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	kafkaProducer, err := kafka.New(cfg.Kafka.Address)
	if err != nil {
		l.Fatal(fmt.Errorf("error creating kafka producer: %w", err))
	}
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

	userUseCase := usecase.NewUserUseCase(
		repo.NewUserRepo(pg),
		kafkaProducer,
	)

	// HTTP Server
	handler := gin.New()

	// Casbin
	csbn := casbin.InitCasbin()

	// New Router
	v1.NewRouter(handler, l, userUseCase, csbn)
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
