package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) GetTasks(ctx context.Context, in service.GetTasksInput) (*service.GetTasksOutput, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GetTasksOutput), args.Error(1)
}

func TestTaskHandler_GetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serializer.SetLogger(zap.NewNop())

	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		queryParams    string
		setup          func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "success - basic request",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20",
			setup: func(svc *MockTaskService) {
				expectedOutput := &service.GetTasksOutput{
					Items: []model.Task{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Status:    "pending",
						},
					},
					HasMore: false,
				}
				svc.On("GetTasks", mock.Anything, mock.MatchedBy(func(in service.GetTasksInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp serializer.Response
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, 0, resp.Code)

				data, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.False(t, data["has_more"].(bool))
				items := data["items"].([]interface{})
				assert.Len(t, items, 1)
			},
		},
		{
			name:           "success - with cursor",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=10&cursor=MTIzNDU2Nzg5MHxhYmNkZWZnaC1pamts",
			setup: func(svc *MockTaskService) {
				expectedOutput := &service.GetTasksOutput{
					Items: []model.Task{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Status:    "success",
						},
					},
					NextCursor: "OTg3NjU0MzIxMHxtbm9wcXJzdC11dnd4",
					HasMore:    true,
				}
				svc.On("GetTasks", mock.Anything, mock.MatchedBy(func(in service.GetTasksInput) bool {
					return in.SessionID == sessionID && in.Limit == 10 && in.Cursor == "MTIzNDU2Nzg5MHxhYmNkZWZnaC1pamts"
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp serializer.Response
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)

				data, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.True(t, data["has_more"].(bool))
				assert.NotEmpty(t, data["next_cursor"])
			},
		},
		{
			name:           "success - using default limit",
			sessionIDParam: sessionID.String(),
			queryParams:    "",
			setup: func(svc *MockTaskService) {
				expectedOutput := &service.GetTasksOutput{
					Items:   []model.Task{},
					HasMore: false,
				}
				svc.On("GetTasks", mock.Anything, mock.MatchedBy(func(in service.GetTasksInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - invalid session id",
			sessionIDParam: "invalid-uuid",
			queryParams:    "?limit=20",
			setup:          func(svc *MockTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit too high",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=300",
			setup:          func(svc *MockTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit too low",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=0",
			setup:          func(svc *MockTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &MockTaskService{}
			tt.setup(svc)

			handler := NewTaskHandler(svc)

			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.GET("/session/:session_id/task", handler.GetTasks)

			req := httptest.NewRequest(http.MethodGet, "/session/"+tt.sessionIDParam+"/task"+tt.queryParams, nil)
			c.Request = req

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			svc.AssertExpectations(t)
		})
	}
}
