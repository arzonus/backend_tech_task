package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PxyUp/backend_tech_task/internal/application"
	application_mock "github.com/PxyUp/backend_tech_task/internal/application/mock"
	"github.com/PxyUp/backend_tech_task/internal/external"
	external_mock "github.com/PxyUp/backend_tech_task/internal/external/mock"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestService_Create(t *testing.T) {
	var objectID = primitive.NewObjectID()
	monkey.Patch(primitive.NewObjectID, func() primitive.ObjectID {
		return objectID
	})

	var now = time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return now
	})

	var cases = map[string]struct {
		UserID string

		ExternalClient_GetExternalStatus_Status external.Status
		ExternalClient_GetExternalStatus_Error  error

		Repository_Create_Error error

		ExpApplication *application.Application
		ExpError       error
	}{
		"success": {
			ExternalClient_GetExternalStatus_Status: external.StatusSkipped,

			UserID: "603bd5e5967f2dba00c8e325",
			ExpApplication: &application.Application{
				ID:             objectID.Hex(),
				Status:         application.StatusOpen,
				UserID:         "603bd5e5967f2dba00c8e325",
				CreatedAt:      now.UTC(),
				ExternalStatus: external.StatusSkipped,
			},
		},
		"failed_invalid_argument": {
			UserID:   "not valid",
			ExpError: fmt.Errorf("invalid argument: user_id is not valid: encoding/hex: invalid byte: U+006E 'n'"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			applicationRepository := application_mock.NewMockRepository(ctrl)
			applicationRepository.
				EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(c.Repository_Create_Error).
				AnyTimes()

			externalClient := external_mock.NewMockClient(ctrl)
			externalClient.
				EXPECT().
				GetExternalStatus(gomock.Any(), gomock.Any()).
				Return(
					c.ExternalClient_GetExternalStatus_Status,
					c.ExternalClient_GetExternalStatus_Error,
				).
				AnyTimes()

			svc := application.NewService(applicationRepository, externalClient)

			app, err := svc.Create(context.Background(), c.UserID)

			if c.ExpError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.ExpError.Error())
			}
			assert.Equal(t, c.ExpApplication, app)
		})
	}
}
