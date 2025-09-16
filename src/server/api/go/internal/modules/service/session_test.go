package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSessionRepo is a mock implementation of SessionRepo
type MockSessionRepo struct {
	mock.Mock
}

func (m *MockSessionRepo) Create(ctx context.Context, s *model.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Delete(ctx context.Context, s *model.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Update(ctx context.Context, s *model.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Get(ctx context.Context, s *model.Session) (*model.Session, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionRepo) CreateMessageWithAssets(ctx context.Context, msg *model.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockSessionRepo) ListBySessionWithCursor(ctx context.Context, sessionID uuid.UUID, afterT time.Time, afterID uuid.UUID, limit int) ([]model.Message, error) {
	args := m.Called(ctx, sessionID, afterT, afterID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Message), args.Error(1)
}

// MockBlobService is a mock implementation of blob service
type MockBlobService struct {
	mock.Mock
}

func (m *MockBlobService) UploadJSON(ctx context.Context, prefix string, data interface{}) (*model.Asset, error) {
	args := m.Called(ctx, prefix, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Asset), args.Error(1)
}

func (m *MockBlobService) DownloadJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockBlobService) PresignGet(ctx context.Context, key string, expire time.Duration) (string, error) {
	args := m.Called(ctx, key, expire)
	return args.String(0), args.Error(1)
}

// MockPublisher is a mock implementation of MQ publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishJSON(ctx context.Context, exchange, routingKey string, data interface{}) error {
	args := m.Called(ctx, exchange, routingKey, data)
	return args.Error(0)
}

func TestSessionService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		session *model.Session
		setup   func(*MockSessionRepo)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful session creation",
			session: &model.Session{
				ID:        uuid.New(),
				ProjectID: uuid.New(),
			},
			setup: func(repo *MockSessionRepo) {
				repo.On("Create", ctx, mock.AnythingOfType("*model.Session")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "creation failure",
			session: &model.Session{
				ID:        uuid.New(),
				ProjectID: uuid.New(),
			},
			setup: func(repo *MockSessionRepo) {
				repo.On("Create", ctx, mock.AnythingOfType("*model.Session")).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockSessionRepo{}
			tt.setup(repo)

			logger := zap.NewNop()
			service := NewSessionService(repo, logger, nil, nil, nil)

			err := service.Create(ctx, tt.session)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestSessionService_Delete(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name      string
		projectID uuid.UUID
		sessionID uuid.UUID
		setup     func(*MockSessionRepo)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful session deletion",
			projectID: projectID,
			sessionID: sessionID,
			setup: func(repo *MockSessionRepo) {
				repo.On("Delete", ctx, mock.MatchedBy(func(s *model.Session) bool {
					return s.ID == sessionID && s.ProjectID == projectID
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "empty session ID",
			projectID: projectID,
			sessionID: uuid.UUID{},
			setup: func(repo *MockSessionRepo) {
				// Empty UUID will call Delete, because len(uuid.UUID{}) != 0
				repo.On("Delete", ctx, mock.AnythingOfType("*model.Session")).Return(nil)
			},
			wantErr: false, // Actually won't error
		},
		{
			name:      "deletion failed",
			projectID: projectID,
			sessionID: sessionID,
			setup: func(repo *MockSessionRepo) {
				repo.On("Delete", ctx, mock.AnythingOfType("*model.Session")).Return(errors.New("deletion failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockSessionRepo{}
			tt.setup(repo)

			logger := zap.NewNop()
			service := NewSessionService(repo, logger, nil, nil, nil)

			err := service.Delete(ctx, tt.projectID, tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestSessionService_GetByID(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()

	tests := []struct {
		name    string
		session *model.Session
		setup   func(*MockSessionRepo)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful session retrieval",
			session: &model.Session{
				ID: sessionID,
			},
			setup: func(repo *MockSessionRepo) {
				expectedSession := &model.Session{
					ID:        sessionID,
					ProjectID: uuid.New(),
				}
				repo.On("Get", ctx, mock.MatchedBy(func(s *model.Session) bool {
					return s.ID == sessionID
				})).Return(expectedSession, nil)
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			session: &model.Session{
				ID: uuid.UUID{},
			},
			setup: func(repo *MockSessionRepo) {
				// Empty UUID will call Get, because len(uuid.UUID{}) != 0
				repo.On("Get", ctx, mock.AnythingOfType("*model.Session")).Return(&model.Session{}, nil)
			},
			wantErr: false,
		},
		{
			name: "retrieval failure",
			session: &model.Session{
				ID: sessionID,
			},
			setup: func(repo *MockSessionRepo) {
				repo.On("Get", ctx, mock.AnythingOfType("*model.Session")).Return(nil, errors.New("session not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockSessionRepo{}
			tt.setup(repo)

			logger := zap.NewNop()
			service := NewSessionService(repo, logger, nil, nil, nil)

			result, err := service.GetByID(ctx, tt.session)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestPartIn_Validate(t *testing.T) {
	tests := []struct {
		name    string
		part    PartIn
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text part",
			part: PartIn{
				Type: "text",
				Text: "This is a piece of text",
			},
			wantErr: false,
		},
		{
			name: "text part with empty text",
			part: PartIn{
				Type: "text",
				Text: "",
			},
			wantErr: true,
			errMsg:  "text part requires non-empty text field",
		},
		{
			name: "valid tool-call part",
			part: PartIn{
				Type: "tool-call",
				Meta: map[string]interface{}{
					"tool_name": "calculator",
					"arguments": map[string]interface{}{
						"expression": "2 + 2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tool-call part missing tool_name",
			part: PartIn{
				Type: "tool-call",
				Meta: map[string]interface{}{
					"arguments": map[string]interface{}{
						"expression": "2 + 2",
					},
				},
			},
			wantErr: true,
			errMsg:  "tool-call part requires 'tool_name' in meta",
		},
		{
			name: "tool-call part missing arguments",
			part: PartIn{
				Type: "tool-call",
				Meta: map[string]interface{}{
					"tool_name": "calculator",
				},
			},
			wantErr: true,
			errMsg:  "tool-call part requires 'arguments' in meta",
		},
		{
			name: "valid tool-result part",
			part: PartIn{
				Type: "tool-result",
				Meta: map[string]interface{}{
					"tool_call_id": "call_123",
					"result":       "4",
				},
			},
			wantErr: false,
		},
		{
			name: "tool-result part missing tool_call_id",
			part: PartIn{
				Type: "tool-result",
				Meta: map[string]interface{}{
					"result": "4",
				},
			},
			wantErr: true,
			errMsg:  "tool-result part requires 'tool_call_id' in meta",
		},
		{
			name: "valid data part",
			part: PartIn{
				Type: "data",
				Meta: map[string]interface{}{
					"data_type": "json",
					"content":   `{"key": "value"}`,
				},
			},
			wantErr: false,
		},
		{
			name: "data part missing data_type",
			part: PartIn{
				Type: "data",
				Meta: map[string]interface{}{
					"content": `{"key": "value"}`,
				},
			},
			wantErr: true,
			errMsg:  "data part requires 'data_type' in meta",
		},
		{
			name: "invalid type",
			part: PartIn{
				Type: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.part.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionService_GetMessages(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()

	tests := []struct {
		name    string
		input   GetMessagesInput
		setup   func(*MockSessionRepo)
		wantErr bool
		errMsg  string
	}{
		{
			name: "repository query failure",
			input: GetMessagesInput{
				SessionID: sessionID,
				Limit:     10,
			},
			setup: func(repo *MockSessionRepo) {
				repo.On("ListBySessionWithCursor", ctx, sessionID, time.Time{}, uuid.UUID{}, 11).Return(nil, errors.New("query failure"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockSessionRepo{}
			tt.setup(repo)

			logger := zap.NewNop()
			service := NewSessionService(repo, logger, nil, nil, nil)

			result, err := service.GetMessages(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
		})
	}
}
