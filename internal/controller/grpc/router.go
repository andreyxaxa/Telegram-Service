package grpc

import (
	v1 "github.com/andreyxaxa/Telegram-Service/internal/controller/grpc/v1"
	"github.com/andreyxaxa/Telegram-Service/internal/usecase"
	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewRouter(app *grpc.Server, t usecase.Telegram, l logger.Interface) {
	{
		v1.NewTelegramRoutes(app, t, l)
	}

	reflection.Register(app)
}
