package usecase

import (
	"context"

	"github.com/andreyxaxa/Telegram-Service/internal/entity"
)

type (
	Telegram interface {
		CreateSession(ctx context.Context) (string, string, error)
		DeleteSession(ctx context.Context, id string) error
		SendMessage(ctx context.Context, sessionID, peer, text string) (int64, error)
		SubscribeMessages(ctx context.Context, sessionID string) (<-chan entity.Message, error)
	}
)
