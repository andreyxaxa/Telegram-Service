package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/andreyxaxa/Telegram-Service/internal/entity"
	"github.com/andreyxaxa/Telegram-Service/pkg/errs"
)

type SessionRepo struct {
	m  map[string]*entity.Session
	mu sync.RWMutex
}

func New() *SessionRepo {
	return &SessionRepo{
		m: make(map[string]*entity.Session),
	}
}

func (r *SessionRepo) Create(_ context.Context, session *entity.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.m[session.ID]; ok {
		return fmt.Errorf("SessionRepo - Create: %w", errs.ErrSessionAlreadyExists)
	}

	r.m[session.ID] = session

	return nil
}

func (r *SessionRepo) GetByID(_ context.Context, id string) (*entity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var s *entity.Session
	var ok bool
	if s, ok = r.m[id]; !ok {
		return nil, fmt.Errorf("SessionRepo - GetByID: %w", errs.ErrSessionNotFound)
	}

	return s, nil
}

func (r *SessionRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// В нашем случае, в usecase.DeleteSession, мы сначала просим GetByID, затем Delete -> получается лишний поход в мапу
	// Поэтому здесь не будем искать в мапе
	// if _, ok := r.m[id]; !ok {
	// 	return fmt.Errorf("SessionRepo - Delete: %w", errs.ErrSessionNotFound)
	// }

	delete(r.m, id)

	return nil
}
