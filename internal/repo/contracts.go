package repo

import (
	"context"

	"github.com/andreyxaxa/Telegram-Service/internal/entity"
)

type (
	SessionRepo interface {
		Create(ctx context.Context, session *entity.Session) error
		GetByID(ctx context.Context, id string) (*entity.Session, error)
		Delete(ctx context.Context, id string) error
	}
)
