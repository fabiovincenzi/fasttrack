//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateAppTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateAppTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateAppTestSuite))
}

func (s *UpdateAppTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name        string
		requestBody request.UpdateApp
	}{
		{
			name: "UpdateApplication",
			requestBody: request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.App
			require.Nil(
				s.T(),
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", app.ID,
				),
			)
			require.Nil(
				s.T(),
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", app.ID,
				),
			)
			assert.Equal(s.T(), "app-type", resp.Type)
			assert.Equal(s.T(), response.AppState{"app-state-key": "new-app-state-value"}, resp.State)
		})
	}
}

func (s *UpdateAppTestSuite) Test_Error() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name        string
		ID          uuid.UUID
		requestBody any
		error       string
	}{
		{
			name: "UpdateAppWithIncorrectState",
			ID:   app.ID,
			requestBody: map[string]any{
				"State": "this-cannot-unmarshal",
			},
			error: "cannot unmarshal",
		},
		{
			name:        "UpdateAppWithUnknownID",
			ID:          uuid.New(),
			requestBody: map[string]any{},
			error:       "Not Found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			require.Nil(
				s.T(),
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", tt.ID,
				),
			)
			assert.Contains(s.T(), resp.Message, tt.error)
		})
	}
}
