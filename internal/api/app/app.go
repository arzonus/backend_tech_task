package app

import (
	"context"

	"github.com/PxyUp/backend_tech_task/internal/api/grpc"
	"github.com/PxyUp/backend_tech_task/internal/api/grpc/services"
	"github.com/PxyUp/backend_tech_task/internal/application"
	application_memory "github.com/PxyUp/backend_tech_task/internal/application/memory"
	application_mongo "github.com/PxyUp/backend_tech_task/internal/application/mongo"
	"github.com/PxyUp/backend_tech_task/internal/external"
	"github.com/PxyUp/backend_tech_task/internal/util/mongoutil"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type App struct {
	grpcServer *grpc.Server
}

func NewApp() (*App, error) {
	log.Logger = log.With().Caller().Logger()

	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}

	mongoDB, err := mongoutil.NewDB(cfg.Mongo)
	if err != nil {
		return nil, err
	}

	externalClient, err := external.NewClient(cfg.External)
	if err != nil {
		return nil, err
	}

	var applicationMongoRepository = application_mongo.NewRepository(mongoDB)
	applicationRepository, err := application_memory.NewRepository(applicationMongoRepository)
	if err != nil {
		return nil, err
	}

	var (
		applicationService = application.NewService(
			applicationRepository,
			externalClient,
		)

		grpcApplicationService = services.NewApplicationService(applicationService)
		grpcServer             = grpc.NewServer(cfg.GRPC, grpcApplicationService)
	)

	return &App{
		grpcServer: grpcServer,
	}, nil
}

func (app App) Run(ctx context.Context) error {
	return app.grpcServer.Run(ctx)
}

type Config struct {
	GRPC     grpc.Config      `envconfig:"grpc"`
	Mongo    mongoutil.Config `envconfig:"mongo"`
	External external.Config  `envconfig:"external"`
}

func NewConfig() (*Config, error) {
	var cfg = new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
