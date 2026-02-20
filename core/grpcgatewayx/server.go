package grpcgatewayx

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	defers []func()
}

func (s *Server) Start() error {
	for _, deferFunc := range s.defers {
		defer deferFunc()
	}
	return s.Server.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}
