package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App    `yaml:"app"`
		HTTP   `yaml:"http"`
		Log    `yaml:"logger"`
		PG     `yaml:"postgres"`
		Kafka  `yaml:"kafka"`
		Stripe `yaml:"stripe"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	// PG -.
	PG struct {
		//PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL         string `env-required:"true"                 env:"PG_URL"`
		TablePrefix string `env:"PG_TABLE_PREFIX" yaml:"table_prefix" env-required:"false"`
	}

	Kafka struct {
		Address string `yaml:"kafka_address"`
	}

	Stripe struct {
		SecretKey string `env:"STRIPE_SECRET_KEY" env-required:"true"`
	}

	// RMQ -.
	//RMQ struct {
	//	ServerExchange string `env-required:"true" yaml:"rpc_server_exchange" env:"RMQ_RPC_SERVER"`
	//	ClientExchange string `env-required:"true" yaml:"rpc_client_exchange" env:"RMQ_RPC_CLIENT"`
	//	URL            string `env-required:"true"                            env:"RMQ_URL"`
	//}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file", err)
	}

	err = cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
