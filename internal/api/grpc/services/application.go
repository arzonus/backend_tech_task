package services

import (
	"context"
	"errors"

	"github.com/PxyUp/backend_tech_task/internal/application"
	api "github.com/PxyUp/backend_tech_task/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ApplicationService struct {
	applicationService application.Service
}

func NewApplicationService(applicationService application.Service) *ApplicationService {
	return &ApplicationService{applicationService: applicationService}
}

var StatusInternal = status.New(codes.Internal, "internal server error")

func (svc ApplicationService) CreateApplication(ctx context.Context, req *api.CreateApplicationRequest) (*api.Application, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	app, err := svc.applicationService.Create(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, application.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, application.ErrApplicationNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, StatusInternal.Err()
	}
	return NewApplication(app), nil
}

func (svc ApplicationService) GetApplicationById(ctx context.Context, req *api.GetApplicationByIdRequest) (*api.Application, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	app, err := svc.applicationService.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, application.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, application.ErrApplicationNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, StatusInternal.Err()
	}
	return NewApplication(app), nil
}

func (svc ApplicationService) GetApplicationsByFilters(ctx context.Context, req *api.GetApplicationsByFiltersRequest) (*api.GetApplicationsByFiltersResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := ParseGetApplicationsByFiltersRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	apps, err := svc.applicationService.GetByFilters(ctx, params)
	if err != nil {
		if errors.Is(err, application.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, StatusInternal.Err()
	}
	return NewGetApplicationsByFiltersResponse(apps), nil
}

func ParseGetApplicationsByFiltersRequest(req *api.GetApplicationsByFiltersRequest) (*application.GetByFilterParams, error) {
	var (
		params = new(application.GetByFilterParams)
		err    error
	)

	params.CreatedAt, err = ParseTimeRange(req.GetCreatedAtTimerange())
	if err != nil {
		return nil, err
	}

	params.UpdatedAt, err = ParseTimeRange(req.GetUpdatedAtTimerange())
	if err != nil {
		return nil, err
	}
	if req.GetStatus() != api.Application_APPLICATION_STATUS_UNSPECIFIED {
		s := ParseApplicationStatus(req.GetStatus())
		params.Status = &s
	}
	if req.GetUserId() != "" {
		userID := req.GetUserId()
		params.UserID = &userID
	}

	return params, nil
}

func NewGetApplicationsByFiltersResponse(apps []application.Application) *api.GetApplicationsByFiltersResponse {
	return &api.GetApplicationsByFiltersResponse{
		Applications: NewApplications(apps),
	}
}

func (svc ApplicationService) UpdateApplication(ctx context.Context, req *api.UpdateApplicationRequest) (*api.Application, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := ParseUpdateApplicationRequest(req)

	app, err := svc.applicationService.Update(ctx, params)
	if err != nil {
		if errors.Is(err, application.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, application.ErrApplicationNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, StatusInternal.Err()
	}
	return NewApplication(app), nil
}

func ParseUpdateApplicationRequest(req *api.UpdateApplicationRequest) *application.UpdateParams {
	return &application.UpdateParams{
		ID:     req.GetId(),
		Status: ParseApplicationStatus(req.GetStatus()),
	}
}
