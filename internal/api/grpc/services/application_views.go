package services

import (
	"github.com/PxyUp/backend_tech_task/internal/application"
	"github.com/PxyUp/backend_tech_task/internal/external"
	api "github.com/PxyUp/backend_tech_task/pkg/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewApplication(app *application.Application) *api.Application {
	if app == nil {
		return nil
	}

	return &api.Application{
		Id:             app.ID,
		Status:         NewApplicationStatus(app.Status),
		UserId:         app.UserID,
		CreatedAt:      timestamppb.New(app.CreatedAt),
		UpdatedAt:      timestamppb.New(app.UpdatedAt),
		ExternalStatus: NewApplicationExternalStatus(app.ExternalStatus),
	}
}

func NewApplicationStatus(status application.Status) api.Application_Status {
	return api.Application_Status(status)
}

func ParseApplicationStatus(status api.Application_Status) application.Status {
	return application.Status(status)
}

func NewApplicationExternalStatus(status external.Status) api.Application_ExternalStatus {
	return api.Application_ExternalStatus(status)
}

func NewApplications(apps []application.Application) []*api.Application {
	if len(apps) == 0 {
		return nil
	}
	var views = make([]*api.Application, len(apps))
	for i := range apps {
		views[i] = NewApplication(&apps[i])
	}
	return views
}

func ParseTimeRange(req *api.TimeRange) (*application.TimeRange, error) {
	if req == nil {
		return nil, nil
	}

	if err := req.Start.CheckValid(); err != nil {
		return nil, err
	}

	if err := req.End.CheckValid(); err != nil {
		return nil, err
	}

	return &application.TimeRange{
		Start: req.Start.AsTime(),
		End:   req.End.AsTime(),
	}, nil
}
