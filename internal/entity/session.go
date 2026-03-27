package entity

import (
	"context"
	"sync/atomic"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
)

type Session struct {
	ID          string
	Client      *telegram.Client
	PeerManager *peers.Manager
	Cancel      context.CancelFunc

	IncomingMessages chan Message
	Authorized       atomic.Bool
}
