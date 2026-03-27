package v1

import (
	v1 "github.com/andreyxaxa/Telegram-Service/docs/proto/v1"
	"github.com/andreyxaxa/Telegram-Service/internal/usecase"
	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
)

type V1 struct {
	v1.UnimplementedTelegramServiceServer

	t usecase.Telegram
	l logger.Interface
}
