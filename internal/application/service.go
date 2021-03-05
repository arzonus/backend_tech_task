package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PxyUp/backend_tech_task/internal/external"
	"go.mongodb.org/mongo-driver/bson/primitive"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate = validator.New()

var (
	ErrInvalidArgument     = fmt.Errorf("invalid argument")
	ErrExternalService     = fmt.Errorf("undefined error from external service")
	ErrRepository          = fmt.Errorf("undefined repository error")
	ErrApplicationNotFound = fmt.Errorf("application is not found")
)

type Service interface {
	Create(ctx context.Context, userID string) (*Application, error)
	GetByID(ctx context.Context, id string) (*Application, error)
	GetByFilters(ctx context.Context, params *GetByFilterParams) ([]Application, error)
	Update(ctx context.Context, params *UpdateParams) (*Application, error)
}

type GetByFilterParams struct {
	Status    *Status    `validate:"required_without_all=UserID CreatedAt UpdatedAt"`
	UserID    *string    `validate:"required_without_all=Status CreatedAt UpdatedAt"`
	CreatedAt *TimeRange `validate:"required_without_all=UserID Status UpdatedAt"`
	UpdatedAt *TimeRange `validate:"required_without_all=UserID Status CreatedAt"`
}

func (p GetByFilterParams) Validate() error {
	if err := validate.Struct(p); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, err.Error())
	}
	return nil
}

type TimeRange struct {
	Start time.Time `validate:"ltfield=End"`
	End   time.Time `validate:"gtfield=Start"`
}

func (tr TimeRange) Validate() error {
	if err := validate.Struct(tr); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, err.Error())
	}
	return nil
}

type UpdateParams struct {
	ID     string
	Status Status
}

func (p UpdateParams) Validate() error {
	if err := validateObjectID(p.ID, "id"); err != nil {
		return err
	}
	if err := p.Status.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, err.Error())
	}
	return nil
}

type service struct {
	repository     Repository
	externalClient external.Client
}

func NewService(
	repository Repository,
	externalClient external.Client,
) Service {
	return &service{
		repository:     repository,
		externalClient: externalClient,
	}
}

func validateObjectID(id string, fieldName string) error {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return fmt.Errorf("%w: %s is not valid: %s", ErrInvalidArgument, fieldName, err.Error())
	}
	return nil
}

func (svc service) Create(ctx context.Context, userID string) (*Application, error) {
	log.Info().Str("user_id", userID).Msg("try to create application")
	if err := validateObjectID(userID, "user_id"); err != nil {
		log.Err(err).Msg("couldn't validate user id ")
		return nil, err
	}

	var app = &Application{
		ID:        primitive.NewObjectID().Hex(),
		Status:    StatusOpen,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	log.Info().Msg("try to get external status")
	externalStatus, err := svc.externalClient.GetExternalStatus(ctx, app.ID)
	if err != nil {
		log.Err(err).Msg("couldn't get external status")
		return nil, fmt.Errorf("%w: %s", ErrExternalService, err.Error())
	}
	app.ExternalStatus = externalStatus

	log.Info().Msgf("got external status: %s, try to save application", app.ExternalStatus.String())
	if err := svc.repository.Create(ctx, app); err != nil {
		log.Err(err).Msgf("couldn't save application")
		return nil, ErrRepository
	}

	log.Info().Msgf("application %s has been created", app.ID)
	return app, nil
}

func (svc service) GetByID(ctx context.Context, id string) (*Application, error) {
	log.Info().Str("id", id).Msg("try to find application")
	if err := validateObjectID(id, "id"); err != nil {
		return nil, err
	}

	log.Info().Msg("try to search application in db")
	app, err := svc.repository.FindByID(ctx, id)
	if err != nil {
		log.Err(err).Msgf("couldn't find application")
		if !errors.Is(err, ErrApplicationNotFound) {
			return nil, ErrRepository
		}
		return nil, err
	}

	log.Err(err).Msgf("application has been found")
	return app, nil
}

func (svc service) GetByFilters(ctx context.Context, params *GetByFilterParams) ([]Application, error) {
	log.Info().Interface("params", params).Msg("try to find applications by filter")
	if err := params.Validate(); err != nil {
		return nil, err
	}

	log.Info().Msg("try to search application by filter in db")
	apps, err := svc.repository.FindByFilters(ctx, params)
	if err != nil {
		log.Err(err).Msgf("couldn't find applications")
		return nil, ErrRepository
	}

	log.Info().Msgf("applications has been found")
	return apps, nil
}

func (svc service) Update(ctx context.Context, params *UpdateParams) (*Application, error) {
	log.Info().Interface("params", params).Msg("try to update application")
	if err := params.Validate(); err != nil {
		return nil, err
	}

	log.Info().Msg("try to find application")
	_, err := svc.GetByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("application has been found, try to update")
	app, err := svc.repository.Update(ctx, params)
	if err != nil {
		log.Err(err).Msgf("couldn't update application")
		if !errors.Is(err, ErrApplicationNotFound) {
			return nil, ErrRepository
		}
		return nil, err
	}

	log.Info().Msg("application has been updated")
	return app, nil
}
