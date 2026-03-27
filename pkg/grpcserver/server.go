package grpcserver

import (
	"context"
	"errors"
	"net"

	"github.com/andreyxaxa/Telegram-Service/pkg/logger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	_defaultAddr = ":80"
)

type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App     *grpc.Server
	notify  chan error
	address string

	logger logger.Interface
}

func New(l logger.Interface, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1)

	s := &Server{
		ctx:     ctx,
		eg:      group,
		App:     grpc.NewServer(),
		notify:  make(chan error, 1),
		address: _defaultAddr,
		logger:  l,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Start() {
	s.eg.Go(func() error {
		var ls net.ListenConfig

		ln, err := ls.Listen(s.ctx, "tcp", s.address)
		if err != nil {
			s.notify <- err
			close(s.notify)

			return err
		}

		err = s.App.Serve(ln)
		if err != nil {
			s.notify <- err
			close(s.notify)

			return err
		}

		return nil
	})

	s.logger.Info("grpc server - Server - Started")
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	var shutdownErrors []error

	s.App.GracefulStop()

	err := s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Error(err, "grpc server - Server - Shutdown - s.eg.Wait")

		shutdownErrors = append(shutdownErrors, err)
	}

	s.logger.Info("grpc server - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}
