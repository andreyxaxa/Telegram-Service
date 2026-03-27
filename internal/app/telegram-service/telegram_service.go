package telegramservice

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreyxaxa/Telegram-Service/config"
	"github.com/andreyxaxa/Telegram-Service/internal/controller/grpc"
	"github.com/andreyxaxa/Telegram-Service/internal/repo/session/inmemory"
	"github.com/andreyxaxa/Telegram-Service/internal/usecase/telegram"
	"github.com/andreyxaxa/Telegram-Service/pkg/grpcserver"
	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
)

func Run(cfg *config.Config) {
	// Logger
	l := logger.New(cfg.Log.Level)

	// Repository
	sessionRepo := inmemory.New()

	// Use-Case
	telegramUC := telegram.New(
		sessionRepo,
		l,
		cfg.TelegramAppCredentials.ID,
		cfg.TelegramAppCredentials.Hash,
		cfg.TelegramService.IncMessagesChanBuffer,
		cfg.TelegramService.QRTokenTimeout,
	)

	// gRPC Server
	grpcServer := grpcserver.New(l, grpcserver.Port(cfg.GRPC.Port))
	grpc.NewRouter(grpcServer.App, telegramUC, l)

	// Start Server
	grpcServer.Start()

	// Waiting Signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("telegram service - Run - signal: %s", s.String())
	case err := <-grpcServer.Notify():
		l.Error(fmt.Errorf("telegram service - Run - grpcServer.Notify: %v", err))
	}

	// Shutdown
	err := grpcServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("telegram service - Run - grpcServer.Shutdown: %v", err))
	}
}
