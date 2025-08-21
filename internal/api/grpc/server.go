package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	pvz "github.com/shrtyk/avito-pvz-test-assignment/proto/pvz/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	wg         *sync.WaitGroup
	port       string
	appService service.Service
	logger     *slog.Logger
	grpcServ   *grpc.Server

	pvz.UnimplementedPVZServiceServer
}

func NewGRPCServer(
	wg *sync.WaitGroup,
	appService service.Service,
	logger *slog.Logger,
	port string,
) *Server {
	s := &Server{
		wg:         wg,
		port:       port,
		appService: appService,
		logger:     logger,
		grpcServ:   grpc.NewServer(),
	}

	pvz.RegisterPVZServiceServer(s.grpcServ, s)
	reflection.Register(s.grpcServ)

	return s
}

func (s *Server) MustStart() {
	l, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		msg := fmt.Sprintf("failed create net.Listener: %s", err)
		panic(msg)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Info("grpc server runnning", slog.String("port", s.port))
		if err := s.grpcServ.Serve(l); err != nil {
			msg := fmt.Sprintf("failed to start grpc server: %s", err)
			panic(msg)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.grpcServ.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.logger.Warn("grpcs server graceful shutdown time out; forcing stop")
		s.grpcServ.Stop()
		return ctx.Err()
	case <-done:
		s.logger.Info("grpc server graceful shutdown complete")
		return nil
	}
}
