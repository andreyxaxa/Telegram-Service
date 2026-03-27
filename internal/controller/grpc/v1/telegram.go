package v1

import (
	"context"
	"errors"

	v1 "github.com/andreyxaxa/Telegram-Service/docs/proto/v1"
	"github.com/andreyxaxa/Telegram-Service/pkg/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *V1) CreateSession(ctx context.Context, _ *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	sessionID, qrStr, err := r.t.CreateSession(ctx)
	if err != nil {
		r.l.Error(err, "grpc - v1 - CreateSession")

		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &v1.CreateSessionResponse{
		SessionId: sessionID,
		QrCode:    qrStr,
	}, nil
}

func (r *V1) DeleteSession(ctx context.Context, req *v1.DeleteSessionRequest) (*v1.DeleteSessionResponse, error) {
	err := r.t.DeleteSession(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, errs.ErrSessionNotFound) {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		r.l.Error(err, "grpc - v1 - DeleteSession")

		return nil, status.Error(codes.Internal, "failed to delete session")
	}

	return &v1.DeleteSessionResponse{}, nil
}

func (r *V1) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageResponse, error) {
	messageID, err := r.t.SendMessage(ctx, req.SessionId, req.Peer, req.Text)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrSessionNotFound):
			return nil, status.Error(codes.NotFound, "session not found")
		case errors.Is(err, errs.ErrSessionNotAuthorized):
			return nil, status.Error(codes.FailedPrecondition, "session not authorized")
		default:
			r.l.Error(err, "grpc - v1 - SendMessage")

			return nil, status.Error(codes.Internal, "failed to send message")
		}
	}

	return &v1.SendMessageResponse{
		MessageId: messageID,
	}, nil
}

func (r *V1) SubscribeMessages(req *v1.SubscribeMessagesRequest, stream grpc.ServerStreamingServer[v1.MessageUpdate]) error {
	msgCh, err := r.t.SubscribeMessages(stream.Context(), req.SessionId)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrSessionNotFound):
			return status.Error(codes.NotFound, "session not found")
		case errors.Is(err, errs.ErrSessionNotAuthorized):
			return status.Error(codes.FailedPrecondition, "session not authorized")
		default:
			r.l.Error(err, "grpc - 1 - SubscribeMessages")

			return status.Error(codes.Internal, "failed to subscribe")
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg, ok := <-msgCh:
			if !ok {
				return status.Error(codes.Unavailable, "session deleted")
			}
			err := stream.Send(&v1.MessageUpdate{
				MessageId: msg.ID,
				From:      msg.From,
				Text:      msg.Text,
				Timestamp: msg.Timestamp,
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send message update: %v", err)
			}
		}
	}
}
