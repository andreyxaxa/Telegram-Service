package peerstorage

import (
	"context"
	"sync"

	"github.com/gotd/td/telegram/peers"
)

type PeerStorage struct {
	peers sync.Map
}

var _ peers.Storage = (*PeerStorage)(nil)

func New() *PeerStorage {
	return &PeerStorage{}
}

func (s *PeerStorage) Save(ctx context.Context, key peers.Key, value peers.Value) error {
	s.peers.Store(key, value)

	return nil
}

func (s *PeerStorage) Find(ctx context.Context, key peers.Key) (value peers.Value, found bool, _ error) {
	v, ok := s.peers.Load(key)
	if !ok {
		return peers.Value{}, false, nil
	}
	return v.(peers.Value), true, nil
}

// stub
func (s *PeerStorage) SavePhone(_ context.Context, _ string, _ peers.Key) error {
	return nil
}

// stub
func (s *PeerStorage) FindPhone(_ context.Context, _ string) (peers.Key, peers.Value, bool, error) {
	return peers.Key{}, peers.Value{}, false, nil
}

// stub
func (s *PeerStorage) GetContactsHash(_ context.Context) (int64, error) {
	return 0, nil
}

// stub
func (s *PeerStorage) SaveContactsHash(_ context.Context, _ int64) error {
	return nil
}
