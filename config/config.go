package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		Log                    Log
		GRPC                   GRPC
		TelegramAppCredentials TelegramAppCredentials
		TelegramService        TelegramService
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	GRPC struct {
		Port string `env:"GRPC_PORT,required"`
	}

	TelegramAppCredentials struct {
		ID   int    `env:"TG_APP_ID,required"`
		Hash string `env:"TG_APP_HASH,required"`
	}

	TelegramService struct {
		IncMessagesChanBuffer int           `env:"TG_SVC_CHAN_BUF" envDefault:"100"`
		QRTokenTimeout        time.Duration `env:"TG_SVC_QR_TIMEOUT" envDefault:"20s"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
