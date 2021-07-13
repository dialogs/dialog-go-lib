package service

import (
	"context"
	"net"
	"sync"
)

type Servable interface {
	Run(lis net.Listener) error
	Stop()
}

type Service struct {
	addr     string
	ctx      context.Context
	servable Servable
	errors   chan error
}

func NewService(ctx context.Context, addr string, servable Servable) *Service {
	return &Service{ctx: ctx, addr: addr, servable: servable, errors: make(chan error, 1)}
}

func (s *Service) GetErrors() chan error {
	return s.errors
}

func (s *Service) Serve(wg *sync.WaitGroup) {
	wg.Add(1)
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.errors <- err
		wg.Done()
		return
	}
	go func() {
		err := s.servable.Run(lis)
		if err != nil {
			s.errors <- err
		}
		wg.Done()
	}()
	go func() {
		<-s.ctx.Done()
		s.servable.Stop()
	}()
}
