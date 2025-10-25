package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// MockSessionService is a mock implementation of SessionService
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) Create(ctx context.Context, s *model.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionService) Delete(ctx context.Context, projectID uuid.UUID, sessionID uuid.UUID) error {
	args := m.Called(ctx, projectID, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) UpdateByID(ctx context.Context, s *model.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionService) GetByID(ctx context.Context, s *model.Session) (*model.Session, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionService) SendMessage(ctx context.Context, in service.SendMessageInput) (*model.Message, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockSessionService) GetMessages(ctx context.Context, in service.GetMessagesInput) (*service.GetMessagesOutput, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GetMessagesOutput), args.Error(1)
}

func (m *MockSessionService) List(ctx context.Context, projectID uuid.UUID, spaceID *uuid.UUID, notConnected bool) ([]model.Session, error) {
	args := m.Called(ctx, projectID, spaceID, notConnected)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Session), args.Error(1)
}

func setupSessionRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestSessionHandler_GetSessions(t *testing.T) {
	projectID := uuid.New()
	spaceID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:        "successful sessions retrieval - all sessions",
			queryParams: "",
			setup: func(svc *MockSessionService) {
				expectedSessions := []model.Session{
					{
						ID:        uuid.New(),
						ProjectID: projectID,
						Configs:   datatypes.JSONMap{"temperature": 0.7},
					},
					{
						ID:        uuid.New(),
						ProjectID: projectID,
						Configs:   datatypes.JSONMap{"model": "gpt-4"},
					},
				}
				svc.On("List", mock.Anything, projectID, (*uuid.UUID)(nil), false).Return(expectedSessions, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful sessions retrieval - filter by space_id",
			queryParams: "?space_id=" + spaceID.String(),
			setup: func(svc *MockSessionService) {
				expectedSessions := []model.Session{
					{
						ID:        uuid.New(),
						ProjectID: projectID,
						SpaceID:   &spaceID,
						Configs:   datatypes.JSONMap{},
					},
				}
				svc.On("List", mock.Anything, projectID, &spaceID, false).Return(expectedSessions, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful sessions retrieval - not connected",
			queryParams: "?not_connected=true",
			setup: func(svc *MockSessionService) {
				expectedSessions := []model.Session{
					{
						ID:        uuid.New(),
						ProjectID: projectID,
						SpaceID:   nil,
						Configs:   datatypes.JSONMap{},
					},
				}
				svc.On("List", mock.Anything, projectID, (*uuid.UUID)(nil), true).Return(expectedSessions, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "empty sessions list",
			queryParams: "",
			setup: func(svc *MockSessionService) {
				svc.On("List", mock.Anything, projectID, (*uuid.UUID)(nil), false).Return([]model.Session{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid space_id",
			queryParams: "?space_id=invalid-uuid",
			setup: func(svc *MockSessionService) {
				// No service call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service layer error",
			queryParams: "",
			setup: func(svc *MockSessionService) {
				svc.On("List", mock.Anything, projectID, (*uuid.UUID)(nil), false).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.GET("/session", func(c *gin.Context) {
				project := &model.Project{ID: projectID}
				c.Set("project", project)
				handler.GetSessions(c)
			})

			req := httptest.NewRequest("GET", "/session"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_CreateSession(t *testing.T) {
	projectID := uuid.New()

	tests := []struct {
		name           string
		requestBody    CreateSessionReq
		setup          func(*MockSessionService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful session creation",
			requestBody: CreateSessionReq{
				Configs: map[string]interface{}{
					"temperature": 0.7,
					"max_tokens":  1000,
				},
			},
			setup: func(svc *MockSessionService) {
				svc.On("Create", mock.Anything, mock.MatchedBy(func(s *model.Session) bool {
					return s.ProjectID == projectID
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "session creation with space ID",
			requestBody: CreateSessionReq{
				SpaceID: uuid.New().String(),
				Configs: map[string]interface{}{
					"model": "gpt-4",
				},
			},
			setup: func(svc *MockSessionService) {
				svc.On("Create", mock.Anything, mock.MatchedBy(func(s *model.Session) bool {
					return s.ProjectID == projectID && s.SpaceID != nil
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "invalid space ID",
			requestBody: CreateSessionReq{
				SpaceID: "invalid-uuid",
				Configs: map[string]interface{}{},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "service layer error",
			requestBody: CreateSessionReq{
				Configs: map[string]interface{}{},
			},
			setup: func(svc *MockSessionService) {
				svc.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.POST("/session", func(c *gin.Context) {
				// Simulate middleware setting project information
				project := &model.Project{ID: projectID}
				c.Set("project", project)
				handler.CreateSession(c)
			})

			body, _ := sonic.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/session", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_DeleteSession(t *testing.T) {
	projectID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful session deletion",
			sessionIDParam: sessionID.String(),
			setup: func(svc *MockSessionService) {
				svc.On("Delete", mock.Anything, projectID, sessionID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			setup: func(svc *MockSessionService) {
				svc.On("Delete", mock.Anything, projectID, sessionID).Return(errors.New("deletion failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.DELETE("/session/:session_id", func(c *gin.Context) {
				project := &model.Project{ID: projectID}
				c.Set("project", project)
				handler.DeleteSession(c)
			})

			req := httptest.NewRequest("DELETE", "/session/"+tt.sessionIDParam, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_UpdateConfigs(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		requestBody    UpdateSessionConfigsReq
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful config update",
			sessionIDParam: sessionID.String(),
			requestBody: UpdateSessionConfigsReq{
				Configs: map[string]interface{}{
					"temperature": 0.8,
					"max_tokens":  2000,
				},
			},
			setup: func(svc *MockSessionService) {
				svc.On("UpdateByID", mock.Anything, mock.MatchedBy(func(s *model.Session) bool {
					return s.ID == sessionID
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			requestBody: UpdateSessionConfigsReq{
				Configs: map[string]interface{}{},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			requestBody: UpdateSessionConfigsReq{
				Configs: map[string]interface{}{},
			},
			setup: func(svc *MockSessionService) {
				svc.On("UpdateByID", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.PUT("/session/:session_id/configs", handler.UpdateConfigs)

			body, _ := sonic.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/session/"+tt.sessionIDParam+"/configs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_GetConfigs(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful config retrieval",
			sessionIDParam: sessionID.String(),
			setup: func(svc *MockSessionService) {
				expectedSession := &model.Session{
					ID:      sessionID,
					Configs: datatypes.JSONMap{"temperature": 0.7},
				}
				svc.On("GetByID", mock.Anything, mock.MatchedBy(func(s *model.Session) bool {
					return s.ID == sessionID
				})).Return(expectedSession, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			setup: func(svc *MockSessionService) {
				svc.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("session not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.GET("/session/:session_id/configs", handler.GetConfigs)

			req := httptest.NewRequest("GET", "/session/"+tt.sessionIDParam+"/configs", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_ConnectToSpace(t *testing.T) {
	sessionID := uuid.New()
	spaceID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		requestBody    ConnectToSpaceReq
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful space connection",
			sessionIDParam: sessionID.String(),
			requestBody: ConnectToSpaceReq{
				SpaceID: spaceID.String(),
			},
			setup: func(svc *MockSessionService) {
				svc.On("UpdateByID", mock.Anything, mock.MatchedBy(func(s *model.Session) bool {
					return s.ID == sessionID && s.SpaceID != nil && *s.SpaceID == spaceID
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			requestBody: ConnectToSpaceReq{
				SpaceID: spaceID.String(),
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid space ID",
			sessionIDParam: sessionID.String(),
			requestBody: ConnectToSpaceReq{
				SpaceID: "invalid-uuid",
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			requestBody: ConnectToSpaceReq{
				SpaceID: spaceID.String(),
			},
			setup: func(svc *MockSessionService) {
				svc.On("UpdateByID", mock.Anything, mock.Anything).Return(errors.New("connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.POST("/session/:session_id/connect_to_space", handler.ConnectToSpace)

			body, _ := sonic.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/session/"+tt.sessionIDParam+"/connect_to_space", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_SendMessage(t *testing.T) {
	projectID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		requestBody    map[string]interface{} // Use map to support different part formats
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		// Acontext format tests
		{
			name:           "acontext format - successful text message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "acontext",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello, world!",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "acontext format - assistant with tool-call",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "acontext",
				"parts": []map[string]interface{}{
					{
						"type": "tool-call",
						"meta": map[string]interface{}{
							"id":        "call_123",
							"tool_name": "get_weather",
							"arguments": map[string]interface{}{"city": "SF"},
						},
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "assistant",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "assistant"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "acontext format - invalid role",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "invalid_role",
				"format": "acontext",
				"parts": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},

		// OpenAI format tests
		{
			name:           "openai format - successful text message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello from OpenAI format!",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "openai format - image_url message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url":    "https://example.com/image.jpg",
							"detail": "high",
						},
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "openai format - tool_call message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "tool_call",
						"id":   "call_abc123",
						"function": map[string]interface{}{
							"name":      "get_weather",
							"arguments": `{"city":"San Francisco"}`,
						},
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "assistant",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "assistant"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "openai format - tool_result message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "tool",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type":         "tool_result",
						"tool_call_id": "call_abc123",
						"output":       "Sunny, 72°F",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user", // tool role converts to user
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "openai format - empty text should fail",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "",
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "openai format - tool_call without ID should fail",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "tool_call",
						"function": map[string]interface{}{
							"name":      "get_weather",
							"arguments": `{"city":"SF"}`,
						},
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},

		// Anthropic format tests
		{
			name:           "anthropic format - successful text message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello from Anthropic format!",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "anthropic format - image message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       "base64data...",
						},
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "anthropic format - tool_use message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "tool_use",
						"id":   "toolu_abc123",
						"name": "get_weather",
						"input": map[string]interface{}{
							"city": "San Francisco",
						},
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "assistant",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "assistant"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "anthropic format - tool_result message",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type":        "tool_result",
						"tool_use_id": "toolu_abc123",
						"content":     "Sunny, 72°F",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "anthropic format - tool_result with error flag",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type":        "tool_result",
						"tool_use_id": "toolu_abc123",
						"content":     "Error: Invalid input",
						"is_error":    true,
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "anthropic format - system role should fail",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "system",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "You are a helpful assistant",
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "anthropic format - empty text should fail",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "",
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "anthropic format - tool_use without ID should fail",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "tool_use",
						"name": "get_weather",
						"input": map[string]interface{}{
							"city": "SF",
						},
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},

		// Default format (OpenAI) tests
		{
			name:           "default format (openai) - text message without format specified",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello, default format!",
					},
				},
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user"
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},

		// Error cases
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			requestBody: map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid format",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "invalid_format",
				"parts": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
			setup: func(svc *MockSessionService) {
				svc.On("SendMessage", mock.Anything, mock.Anything).Return(nil, errors.New("send failed"))
			},
			expectedStatus: http.StatusBadRequest,
		},

		// Additional edge cases and error scenarios
		{
			name:           "missing role field",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"parts": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing parts field",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role": "user",
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty parts array",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON structure",
			sessionIDParam: sessionID.String(),
			requestBody:    map[string]interface{}{"invalid": "json structure"},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "acontext format - missing required fields in tool-call",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "acontext",
				"parts": []map[string]interface{}{
					{
						"type": "tool-call",
						// missing meta field
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "openai format - missing required fields in tool_call",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type": "tool_call",
						// missing id and function fields
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "anthropic format - missing required fields in tool_use",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "assistant",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type": "tool_use",
						// missing id, name, and input fields
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "openai format - invalid image_url structure",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type":      "image_url",
						"image_url": "invalid_url_structure", // should be an object
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "anthropic format - invalid image source structure",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type":   "image",
						"source": "invalid_source_structure", // should be an object
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "openai format - tool_result without tool_call_id",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "tool",
				"format": "openai",
				"parts": []map[string]interface{}{
					{
						"type":   "tool_result",
						"output": "Result",
						// missing tool_call_id
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "anthropic format - tool_result without tool_use_id",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":   "user",
				"format": "anthropic",
				"parts": []map[string]interface{}{
					{
						"type":    "tool_result",
						"content": "Result",
						// missing tool_use_id
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unsupported part type",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"type": "unsupported_type",
						"text": "Hello",
					},
				},
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed parts structure",
			sessionIDParam: sessionID.String(),
			requestBody: map[string]interface{}{
				"role":  "user",
				"parts": "not_an_array",
			},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.POST("/session/:session_id/messages", func(c *gin.Context) {
				project := &model.Project{ID: projectID}
				c.Set("project", project)
				handler.SendMessage(c)
			})

			body, _ := sonic.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/session/"+tt.sessionIDParam+"/messages", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_GetMessages(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		queryParams    string
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful message retrieval",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session ID",
			sessionIDParam: "invalid-uuid",
			queryParams:    "?limit=20",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit parameter",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=0",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service layer error",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20",
			setup: func(svc *MockSessionService) {
				svc.On("GetMessages", mock.Anything, mock.Anything).Return(nil, errors.New("retrieval failed"))
			},
			expectedStatus: http.StatusBadRequest,
		},

		// Additional edge cases and error scenarios for GetMessages
		{
			name:           "limit exceeds maximum (201)",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=201",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative limit",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=-1",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero limit",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=0",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit format (non-numeric)",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=abc",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid format parameter",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&format=invalid_format",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "with_asset_public_url with invalid boolean",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&with_asset_public_url=maybe",
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "acontext format conversion",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&format=acontext",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "anthropic format conversion",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&format=anthropic",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "pagination with cursor",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&cursor=eyJpZCI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTQyNjYxNDE3NDAwMCJ9",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore:    true,
					NextCursor: "eyJpZCI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTQyNjYxNDE3NDAwMSJ9",
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.Cursor == "eyJpZCI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTQyNjYxNDE3NDAwMCJ9"
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with_asset_public_url false",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&with_asset_public_url=false",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.WithAssetPublicURL == false
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with_asset_public_url true (default)",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&with_asset_public_url=true",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.WithAssetPublicURL == true
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "default limit when not specified",
			sessionIDParam: sessionID.String(),
			queryParams:    "",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 // default limit
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty messages list",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items:   []model.Message{},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "time_desc=false (default)",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&time_desc=false",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.TimeDesc == false
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "time_desc=true",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&time_desc=true",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.TimeDesc == true
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "time_desc with cursor",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&cursor=eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQwMDowMDowMFoiLCJpZCI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTQyNjYxNDE3NDAwMCJ9&time_desc=false",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore:    true,
					NextCursor: "eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQwMDowMDowMFoiLCJpZCI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTQyNjYxNDE3NDAwMSJ9",
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.TimeDesc == false && in.Cursor != ""
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "time_desc with format conversion",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&time_desc=true&format=acontext",
			setup: func(svc *MockSessionService) {
				expectedOutput := &service.GetMessagesOutput{
					Items: []model.Message{
						{
							ID:        uuid.New(),
							SessionID: sessionID,
							Role:      "user",
						},
					},
					HasMore: false,
				}
				svc.On("GetMessages", mock.Anything, mock.MatchedBy(func(in service.GetMessagesInput) bool {
					return in.SessionID == sessionID && in.Limit == 20 && in.TimeDesc == true
				})).Return(expectedOutput, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid time_desc parameter",
			sessionIDParam: sessionID.String(),
			queryParams:    "?limit=20&time_desc=invalid",
			setup: func(svc *MockSessionService) {
				// No service call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.GET("/session/:session_id/messages", handler.GetMessages)

			req := httptest.NewRequest("GET", "/session/"+tt.sessionIDParam+"/messages"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
func TestSessionHandler_SendMessage_Multipart(t *testing.T) {
	projectID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name           string
		sessionIDParam string
		payload        string
		files          map[string]string // field name -> file content
		setup          func(*MockSessionService)
		expectedStatus int
	}{
		{
			name:           "successful multipart message with file",
			sessionIDParam: sessionID.String(),
			payload: `{
				"role": "user",
				"format": "openai",
				"parts": [
					{
						"type": "text",
						"text": "Please analyze this file"
					},
					{
						"type": "image_url",
						"image_url": {
							"url": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
						},
						"file_field": "image_file"
					}
				]
			}`,
			files: map[string]string{
				"image_file": "fake image content",
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user" && len(in.Files) > 0
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "multipart with invalid JSON payload",
			sessionIDParam: sessionID.String(),
			payload:        "invalid json",
			files:          map[string]string{},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "multipart with missing required file",
			sessionIDParam: sessionID.String(),
			payload: `{
				"role": "user",
				"format": "openai",
				"parts": [
					{
						"type": "image_url",
						"image_url": {
							"url": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
						},
						"file_field": "missing_file"
					}
				]
			}`,
			files:          map[string]string{},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "multipart with empty payload",
			sessionIDParam: sessionID.String(),
			payload:        "",
			files:          map[string]string{},
			setup:          func(svc *MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "multipart with acontext format and file",
			sessionIDParam: sessionID.String(),
			payload: `{
				"role": "user",
				"format": "acontext",
				"parts": [
					{
						"type": "text",
						"text": "Please analyze this file"
					},
					{
						"type": "image",
						"file_field": "document_file"
					}
				]
			}`,
			files: map[string]string{
				"document_file": "fake document content",
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user" && len(in.Files) > 0
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "multipart with anthropic format and file",
			sessionIDParam: sessionID.String(),
			payload: `{
				"role": "user",
				"format": "anthropic",
				"parts": [
					{
						"type": "text",
						"text": "Please analyze this file"
					},
					{
						"type": "image",
						"source": {
							"type": "base64",
							"media_type": "image/jpeg",
							"data": "base64data..."
						},
						"file_field": "image_file"
					}
				]
			}`,
			files: map[string]string{
				"image_file": "fake image content",
			},
			setup: func(svc *MockSessionService) {
				expectedMessage := &model.Message{
					ID:        uuid.New(),
					SessionID: sessionID,
					Role:      "user",
				}
				svc.On("SendMessage", mock.Anything, mock.MatchedBy(func(in service.SendMessageInput) bool {
					return in.ProjectID == projectID && in.SessionID == sessionID && in.Role == "user" && len(in.Files) > 0
				})).Return(expectedMessage, nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSessionService{}
			tt.setup(mockService)

			handler := NewSessionHandler(mockService)
			router := setupSessionRouter()
			router.POST("/session/:session_id/messages", func(c *gin.Context) {
				project := &model.Project{ID: projectID}
				c.Set("project", project)
				handler.SendMessage(c)
			})

			// Create multipart form data
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add payload field
			if tt.payload != "" {
				payloadField, _ := writer.CreateFormField("payload")
				payloadField.Write([]byte(tt.payload))
			}

			// Add files
			for fieldName, content := range tt.files {
				fileField, _ := writer.CreateFormFile(fieldName, "test_file.txt")
				fileField.Write([]byte(content))
			}

			writer.Close()

			req := httptest.NewRequest("POST", "/session/"+tt.sessionIDParam+"/messages", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSessionHandler_SendMessage_InvalidJSON(t *testing.T) {
	projectID := uuid.New()
	sessionID := uuid.New()

	t.Run("invalid JSON in request body", func(t *testing.T) {
		mockService := &MockSessionService{}
		// No setup needed as the request should fail before reaching the service

		handler := NewSessionHandler(mockService)
		router := setupSessionRouter()
		router.POST("/session/:session_id/messages", func(c *gin.Context) {
			project := &model.Project{ID: projectID}
			c.Set("project", project)
			handler.SendMessage(c)
		})

		// Send invalid JSON directly
		req := httptest.NewRequest("POST", "/session/"+sessionID.String()+"/messages", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}
