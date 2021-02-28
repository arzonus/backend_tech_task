package application

import (
	"context"
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

//go:generate mockgen -destination=mock/repository.go -package=application_mock "github.com/PxyUp/backend_tech_task/internal/application" Repository
type Repository interface {
	Create(ctx context.Context, application *Application) error
	FindByID(ctx context.Context, id string) (*Application, error)
	FindByFilters(ctx context.Context, filter *GetByFilterParams) ([]Application, error)
	Update(ctx context.Context, params *UpdateParams) (*Application, error)
}

type Config struct {
	Expiration    time.Duration `envconfig:"expiration"`
	PurgeInterval time.Duration `envconfig:"purge_interval"`
}

type repository struct {
	Repository
	cache *cache.Cache

	// using for avoiding multiple changing in one moment
	sg singleflight.Group
}

func NewRepository(cfg Config, dbRepository Repository) Repository {
	if cfg.Expiration == 0 {
		cfg.Expiration = 5 * time.Minute
	}
	if cfg.PurgeInterval == 0 {
		cfg.PurgeInterval = 10 * time.Minute
	}
	return &repository{
		cache:      cache.New(cfg.Expiration, cfg.PurgeInterval),
		Repository: dbRepository,
	}
}

func (r *repository) Update(ctx context.Context, params *UpdateParams) (*Application, error) {
	v, err, _ := r.sg.Do(params.ID, func() (interface{}, error) {
		app, err := r.Repository.Update(ctx, params)
		if err != nil {
			return nil, err
		}

		r.cache.SetDefault(params.ID, app)
		return app, nil
	})
	if err != nil {
		return nil, err
	}

	app, ok := v.(*Application)
	if !ok {
		return nil, fmt.Errorf("update couldn't return app")
	}

	return app, nil
}

func (r *repository) Create(ctx context.Context, app *Application) error {
	_, err, _ := r.sg.Do(app.ID, func() (interface{}, error) {
		if err := r.Repository.Create(ctx, app); err != nil {
			return nil, err
		}
		r.cache.SetDefault(app.ID, app)
		return nil, nil
	})

	return err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Application, error) {
	v, ok := r.cache.Get(id)
	if !ok {
		return r.Repository.FindByID(ctx, id)
	}
	app, ok := v.(*Application)
	if !ok {
		r.cache.Delete(id)
		return r.Repository.FindByID(ctx, id)
	}
	return app, nil
}
