package application

import (
	"context"
)

//go:generate mockgen -destination=mock/repository.go -package=application_mock "github.com/PxyUp/backend_tech_task/internal/application" Repository
type Repository interface {
	Create(ctx context.Context, application *Application) error
	Update(ctx context.Context, params *UpdateParams) (*Application, error)
	FindByID(ctx context.Context, id string) (*Application, error)
	FindByFilters(ctx context.Context, filter *GetByFilterParams) ([]Application, error)
	FindAll(ctx context.Context) ([]Application, error)
}
