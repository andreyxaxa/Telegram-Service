package v1

import (
	v1 "github.com/andreyxaxa/Telegram-Service/docs/proto/v1"
	"github.com/andreyxaxa/Telegram-Service/internal/usecase"
	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
	"google.golang.org/grpc"
)

func NewTelegramRoutes(app *grpc.Server, t usecase.Telegram, l logger.Interface) {
	r := &V1{t: t, l: l}

	{
		v1.RegisterTelegramServiceServer(app, r)
	}
}
