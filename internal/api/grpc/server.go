package grpc

import (
	"github.com/PxyUp/backend_tech_task/internal/api/grpc/services"
	api "github.com/PxyUp/backend_tech_task/pkg/proto"

	"context"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Config struct {
	Address string `envconfig:"address"`
}

type Server struct {
	cfg Config

	applicationService *services.ApplicationService
}

func NewServer(
	cfg Config,
	applicationService *services.ApplicationService,
) *Server {
	if cfg.Address == "" {
		cfg.Address = ":8080"
	}

	return &Server{
		cfg:                cfg,
		applicationService: applicationService,
	}
}

func (srv Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", srv.cfg.Address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	api.RegisterApplicationServiceServer(grpcServer, srv.applicationService)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		_ = listener.Close()
	}()

	log.Info().Msg("grpc server started")
	return grpcServer.Serve(listener)
}
