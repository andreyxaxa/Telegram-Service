package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/andreyxaxa/Telegram-Service/internal/entity"
	"github.com/andreyxaxa/Telegram-Service/internal/repo"
	"github.com/andreyxaxa/Telegram-Service/pkg/errs"
	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
	"github.com/andreyxaxa/Telegram-Service/pkg/peerstorage"
	"github.com/google/uuid"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

type UseCase struct {
	r      repo.SessionRepo
	logger logger.Interface

	appID   int
	appHash string

	chBuf     int
	qrTimeout time.Duration
}

func New(r repo.SessionRepo, l logger.Interface, appID int, appHash string, chBuf int, qrTimeout time.Duration) *UseCase {
	return &UseCase{
		r:         r,
		logger:    l,
		appID:     appID,
		appHash:   appHash,
		chBuf:     chBuf,
		qrTimeout: qrTimeout,
	}
}

func (uc *UseCase) CreateSession(ctx context.Context) (string, string, error) {
	// 1. Генерируем UUID для сессии
	sessionID := uuid.New().String()

	// 2. Создаем канал для qr-строки
	qrCh := make(chan string, 1)

	// 3. Создаем хранилище MTProto сессий
	strg := new(session.StorageMemory)

	// 4. Создаем тг-диспатчера
	dispatcher := tg.NewUpdateDispatcher()

	// 5. Создаем нового клиента
	client := telegram.NewClient(uc.appID, uc.appHash, telegram.Options{
		SessionStorage: strg,
		UpdateHandler:  dispatcher,
	})

	// 6. Создаем сессию(из нашей бизнес-логики)
	sessionCtx, sessionCancel := context.WithCancel(context.Background())

	s := &entity.Session{
		ID:               sessionID,
		Client:           client,
		Cancel:           sessionCancel,
		IncomingMessages: make(chan entity.Message, uc.chBuf),
	}

	// 7. Регистрируем хендлер, он будет вызываться при новых сообщениях
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out { // если исходящее сообщение
			return nil
		}

		var from string
		if peer, ok := msg.FromID.(*tg.PeerUser); ok {
			if user, ok := e.Users[peer.UserID]; ok {
				from = user.Username
			}
		}

		// кладем сообщение в канал
		s.IncomingMessages <- entity.Message{
			ID:        int64(msg.ID),
			From:      from,
			Text:      msg.Message,
			Timestamp: int64(msg.Date),
		}

		return nil
	})

	// 8. Кладем сессию в наше хранилище
	err := uc.r.Create(ctx, s)
	if err != nil {
		// отменяем контекст
		sessionCancel()

		return "", "", fmt.Errorf("UseCase - CreateSession - uc.r.Create: %w", err)
	}

	go func() {
		defer sessionCancel()

		err := client.Run(sessionCtx, func(ctx context.Context) error {
			// создадим хранилище peer'ов для сессии
			peerstrg := peerstorage.New()
			s.PeerManager = peers.Options{
				Storage: peerstrg,
			}.Build(s.Client.API())

			// 1. Регистрируем обработчик на события, связанные с авторизацией
			dispatcher.OnLoginToken(func(ctx context.Context, e tg.Entities, update *tg.UpdateLoginToken) error {
				result, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
					APIID:     uc.appID,
					APIHash:   uc.appHash,
					ExceptIDs: []int64{},
				})
				if err != nil {
					return fmt.Errorf("UseCase - CreateSession - client.Run - client.API().AuthExportLoginToken: %w", err)
				}

				switch result.(type) {
				case *tg.AuthLoginTokenSuccess:
					// успешная авторизация
					s.Authorized.Store(true)
					uc.logger.Info("session authorized, session ID: %s", sessionID)
				default:
					return fmt.Errorf("UseCase - CreateSession - client.Run: %w", errs.ErrUnexpectedTokenType)
				}

				return nil
			})

			// 2. Запрашиваем qr-токен
			result, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
				APIID:     uc.appID,
				APIHash:   uc.appHash,
				ExceptIDs: []int64{},
			})
			if err != nil {
				return fmt.Errorf("UseCase - CreateSession - client.Run - client.API().AuthExportLoginToken: %w", err)
			}

			token, ok := result.(*tg.AuthLoginToken)
			if !ok {
				return fmt.Errorf("UseCase - CreateSession - client.Run: %w", errs.ErrUnexpectedTokenType)
			}

			// 3. Формируем url - из него клиент нарисует qr
			url := "tg://login?token=" + base64.RawURLEncoding.EncodeToString(token.Token)
			qrCh <- url

			// горутина будет висеть здесь, пока не вызовут DeleteSession
			// а зарегестрированные обработчики будут реагировать на интересующие нас события
			<-sessionCtx.Done()

			return nil
		})

		if err != nil {
			uc.logger.Error(err, "UseCase - CreateSession - client.Run")
		}
	}()

	// ждем qr в основной горутине
	select {
	case token := <-qrCh:
		return sessionID, token, nil
	case <-time.After(uc.qrTimeout):
		sessionCancel()
		return "", "", fmt.Errorf("UseCase - CreateSession: %w", errs.ErrQRTokenWaitingTimeout)
	}
}

func (uc *UseCase) DeleteSession(ctx context.Context, id string) error {
	sessionToDelete, err := uc.r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("UseCase - DeleteSession - uc.r.GetByID: %w", err)
	}

	// логаут
	if sessionToDelete.Authorized.Load() {
		_, err := sessionToDelete.Client.API().AuthLogOut(ctx)
		if err != nil {
			uc.logger.Error(err, "UseCase - DeleteSession - sessionToDelete.Client.API().AuthLogOut: %w, err")
		}
	}

	// останавливаем клиент
	sessionToDelete.Cancel()
	// сообщаем, что сообщений больше не будет
	close(sessionToDelete.IncomingMessages)

	// удаляем из repo
	err = uc.r.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("UseCase - DeleteSession - uc.r.Delete: %w", err)
	}

	return nil
}

func (uc *UseCase) SendMessage(ctx context.Context, sessionID, peer, text string) (int64, error) {
	// получаем сессию
	s, err := uc.r.GetByID(ctx, sessionID)
	if err != nil {
		return 0, fmt.Errorf("UseCase - SendMessage - uc.r.GetByID: %w", err)
	}

	// проверяем, авторизована ли
	if !s.Authorized.Load() {
		return 0, fmt.Errorf("UseCase - SendMessage: %w", errs.ErrSessionNotAuthorized)
	}

	// резолвим peer нашим менеджером
	resolverPeer, err := s.PeerManager.Resolve(ctx, peer)
	if err != nil {
		return 0, fmt.Errorf("UseCase - SendMessage - s.PeerManager.Resolve: %w", err)
	}

	sender := message.NewSender(s.Client.API())

	// отправляем сообщение
	res, err := sender.To(resolverPeer.InputPeer()).Text(ctx, text)
	if err != nil {
		return 0, fmt.Errorf("UseCase - SendMessage - sender.To.Text: %w", err)
	}

	switch u := res.(type) {
	case *tg.UpdateShortSentMessage:
		return int64(u.ID), nil
	case *tg.Updates:
		for _, update := range u.Updates {
			switch v := update.(type) {
			case *tg.UpdateMessageID:
				return int64(v.ID), nil
			case *tg.UpdateNewMessage:
				if m, ok := v.Message.(*tg.Message); ok {
					return int64(m.ID), nil
				}

			default:
			}
		}
	default:
	}

	return 0, fmt.Errorf("UseCase - SendMessage: %w", errs.ErrUnexpectedUpdatesType)
}

func (uc *UseCase) SubscribeMessages(ctx context.Context, sessionID string) (<-chan entity.Message, error) {
	s, err := uc.r.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - SubscribeMessages - uc.r.GetByID: %w", err)
	}

	if !s.Authorized.Load() {
		return nil, fmt.Errorf("UseCase - SubscribeMessages: %w", errs.ErrSessionNotAuthorized)
	}

	return s.IncomingMessages, nil
}
