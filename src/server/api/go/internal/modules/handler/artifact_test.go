package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/pkg/utils/fileparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// MockArtifactService is a mock implementation of ArtifactService
type MockArtifactService struct {
	mock.Mock
}

func (m *MockArtifactService) Create(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, userMeta map[string]interface{}) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, path, filename, fileHeader, userMeta)
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) Delete(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) error {
	args := m.Called(ctx, diskID, artifactID)
	return args.Error(0)
}

func (m *MockArtifactService) GetByID(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, artifactID)
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetPresignedURL(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, expire time.Duration) (string, error) {
	args := m.Called(ctx, diskID, artifactID, expire)
	return args.String(0), args.Error(1)
}

func (m *MockArtifactService) UpdateArtifact(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, artifactID, fileHeader, newPath, newFilename)
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) ListByPath(ctx context.Context, diskID uuid.UUID, path string) ([]*model.Artifact, error) {
	args := m.Called(ctx, diskID, path)
	return args.Get(0).([]*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetAllPaths(ctx context.Context, diskID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, diskID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockArtifactService) GetByDiskID(ctx context.Context, diskID uuid.UUID) ([]*model.Artifact, error) {
	args := m.Called(ctx, diskID)
	return args.Get(0).([]*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) DeleteByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) error {
	args := m.Called(ctx, diskID, path, filename)
	return args.Error(0)
}

func (m *MockArtifactService) GetByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, path, filename)
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetPresignedURLByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, expire time.Duration) (string, error) {
	args := m.Called(ctx, diskID, path, filename, expire)
	return args.String(0), args.Error(1)
}

func (m *MockArtifactService) UpdateArtifactByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, path, filename, fileHeader, newPath, newFilename)
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetFileContent(ctx context.Context, diskID uuid.UUID, path string, filename string) (*fileparser.FileContent, error) {
	args := m.Called(ctx, diskID, path, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fileparser.FileContent), args.Error(1)
}

func TestArtifactHandler_UpsertArtifact(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		diskID         string
		filePath       string
		meta           string
		fileContent    string
		fileName       string
		mockSetup      func(*MockArtifactService, string)
		expectedStatus int
	}{
		{
			name:        "successful file upsert",
			diskID:      uuid.New().String(),
			filePath:    "/test/test.txt",
			meta:        `{"description": "test file"}`,
			fileContent: "test content",
			fileName:    "test.txt",
			mockSetup: func(m *MockArtifactService, diskIDStr string) {
				diskID := uuid.MustParse(diskIDStr)
				expectedFile := &model.Artifact{
					ID:       uuid.New(),
					DiskID:   diskID,
					Path:     "/test/",
					Filename: "test.txt",
					Meta: map[string]interface{}{
						model.ArtifactInfoKey: map[string]interface{}{
							"path":     "/test/",
							"filename": "test.txt",
							"mime":     "text/plain",
							"size":     12,
						},
						"description": "test file",
					},
					AssetMeta: datatypes.NewJSONType(model.Asset{
						Bucket: "test-bucket",
						S3Key:  "test-key",
						ETag:   "test-etag",
						SHA256: "test-sha256",
						MIME:   "text/plain",
						SizeB:  12,
					}),
				}
				m.On("Create", mock.Anything, diskID, "/test/", "test.txt", mock.Anything, mock.Anything).Return(expectedFile, nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockArtifactService)
			tt.mockSetup(mockService, tt.diskID)

			handler := NewArtifactHandler(mockService)

			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add file
			fileWriter, err := writer.CreateFormFile("file", tt.fileName)
			assert.NoError(t, err)
			_, err = fileWriter.Write([]byte(tt.fileContent))
			assert.NoError(t, err)

			// Add form fields
			if tt.filePath != "" {
				writer.WriteField("file_path", tt.filePath)
			}
			if tt.meta != "" {
				writer.WriteField("meta", tt.meta)
			}

			writer.Close()

			// Create request
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/disk/%s/artifact", tt.diskID), body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "disk_id", Value: tt.diskID},
			}

			// Call handler
			handler.UpsertArtifact(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response serializer.Response
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestArtifactHandler_DeleteArtifact(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		diskID         string
		filePath       string
		mockSetup      func(*MockArtifactService, string, string)
		expectedStatus int
	}{
		{
			name:     "successful file deletion",
			diskID:   uuid.New().String(),
			filePath: "/test/test.txt",
			mockSetup: func(m *MockArtifactService, diskIDStr string, filePath string) {
				diskID := uuid.MustParse(diskIDStr)
				m.On("DeleteByPath", mock.Anything, diskID, "/test/", "test.txt").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockArtifactService)
			tt.mockSetup(mockService, tt.diskID, tt.filePath)

			handler := NewArtifactHandler(mockService)

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/disk/%s/artifact?file_path=%s", tt.diskID, tt.filePath), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "disk_id", Value: tt.diskID},
			}

			// Call handler
			handler.DeleteArtifact(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockService.AssertExpectations(t)
		})
	}
}

func TestArtifactHandler_UpdateArtifact(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		diskID         string
		filePath       string
		fileContent    string
		fileName       string
		mockSetup      func(m *MockArtifactService, diskIDStr string)
		expectedStatus int
	}{
		{
			name:        "successful file update with same filename",
			diskID:      uuid.New().String(),
			filePath:    "/test/report.pdf",
			fileContent: "updated content",
			fileName:    "report.pdf", // Same filename as in filePath
			mockSetup: func(m *MockArtifactService, diskIDStr string) {
				diskID := uuid.MustParse(diskIDStr)
				expectedFile := &model.Artifact{
					ID:       uuid.New(),
					DiskID:   diskID,
					Path:     "/test/",
					Filename: "report.pdf",
					Meta: map[string]interface{}{
						model.ArtifactInfoKey: map[string]interface{}{
							"path":     "/test/",
							"filename": "report.pdf",
							"mime":     "application/pdf",
							"size":     15,
						},
					},
					AssetMeta: datatypes.NewJSONType(model.Asset{
						Bucket: "test-bucket",
						S3Key:  "test-key",
						ETag:   "test-etag",
						SHA256: "test-sha256",
						MIME:   "application/pdf",
						SizeB:  15,
					}),
				}
				m.On("UpdateArtifactByPath", mock.Anything, diskID, "/test/", "report.pdf", mock.Anything, (*string)(nil), (*string)(nil)).Return(expectedFile, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful file update with different filename",
			diskID:      uuid.New().String(),
			filePath:    "/test/report.pdf",
			fileContent: "updated content",
			fileName:    "new-report.pdf", // Different filename
			mockSetup: func(m *MockArtifactService, diskIDStr string) {
				diskID := uuid.MustParse(diskIDStr)
				expectedFile := &model.Artifact{
					ID:       uuid.New(),
					DiskID:   diskID,
					Path:     "/test/",
					Filename: "new-report.pdf",
					Meta: map[string]interface{}{
						model.ArtifactInfoKey: map[string]interface{}{
							"path":     "/test/",
							"filename": "new-report.pdf",
							"mime":     "application/pdf",
							"size":     15,
						},
					},
					AssetMeta: datatypes.NewJSONType(model.Asset{
						Bucket: "test-bucket",
						S3Key:  "test-key",
						ETag:   "test-etag",
						SHA256: "test-sha256",
						MIME:   "application/pdf",
						SizeB:  15,
					}),
				}
				newFilename := "new-report.pdf"
				m.On("UpdateArtifactByPath", mock.Anything, diskID, "/test/", "report.pdf", mock.Anything, (*string)(nil), &newFilename).Return(expectedFile, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "file update with invalid artifact ID",
			diskID:      "invalid-uuid",
			filePath:    "/test/report.pdf",
			fileContent: "updated content",
			fileName:    "report.pdf",
			mockSetup: func(m *MockArtifactService, diskIDStr string) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "file update with invalid path",
			diskID:      uuid.New().String(),
			filePath:    "/test/../../../report.pdf", // Path traversal attempt
			fileContent: "updated content",
			fileName:    "report.pdf",
			mockSetup: func(m *MockArtifactService, diskIDStr string) {
				// No mock setup needed for invalid path
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockArtifactService)
			tt.mockSetup(mockService, tt.diskID)

			handler := NewArtifactHandler(mockService)

			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add file
			fileWriter, err := writer.CreateFormFile("file", tt.fileName)
			assert.NoError(t, err)
			_, err = fileWriter.Write([]byte(tt.fileContent))
			assert.NoError(t, err)

			// Add form fields
			if tt.filePath != "" {
				writer.WriteField("file_path", tt.filePath)
			}

			writer.Close()

			// Create request
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/disk/%s/artifact", tt.diskID), body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "disk_id", Value: tt.diskID},
			}

			// Call handler
			handler.UpdateArtifact(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response serializer.Response
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestArtifactHandler_GetArtifact(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		diskID         string
		filePath       string
		withContent    bool
		withPublicURL  bool
		mockSetup      func(*MockArtifactService, string, string)
		expectedStatus int
	}{
		{
			name:          "successful artifact retrieval with content",
			diskID:        uuid.New().String(),
			filePath:      "/test/data.csv",
			withContent:   true,
			withPublicURL: true,
			mockSetup: func(m *MockArtifactService, diskIDStr string, filePath string) {
				diskID := uuid.MustParse(diskIDStr)
				expectedFile := &model.Artifact{
					ID:       uuid.New(),
					DiskID:   diskID,
					Path:     "/test/",
					Filename: "data.csv",
					Meta: map[string]interface{}{
						model.ArtifactInfoKey: map[string]interface{}{
							"path":     "/test/",
							"filename": "data.csv",
							"mime":     "text/csv",
							"size":     1024,
						},
					},
					AssetMeta: datatypes.NewJSONType(model.Asset{
						Bucket: "test-bucket",
						S3Key:  "test-key",
						ETag:   "test-etag",
						SHA256: "test-sha256",
						MIME:   "text/csv",
						SizeB:  1024,
					}),
				}
				expectedContent := &fileparser.FileContent{
					Type: "csv",
					Raw:  "name,age\nJohn,25",
				}
				m.On("GetByPath", mock.Anything, diskID, "/test/", "data.csv").Return(expectedFile, nil)
				m.On("GetPresignedURLByPath", mock.Anything, diskID, "/test/", "data.csv", mock.AnythingOfType("time.Duration")).Return("https://example.com/presigned-url", nil)
				m.On("GetFileContent", mock.Anything, diskID, "/test/", "data.csv").Return(expectedContent, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "successful artifact retrieval without content",
			diskID:        uuid.New().String(),
			filePath:      "/test/data.csv",
			withContent:   false,
			withPublicURL: false,
			mockSetup: func(m *MockArtifactService, diskIDStr string, filePath string) {
				diskID := uuid.MustParse(diskIDStr)
				expectedFile := &model.Artifact{
					ID:       uuid.New(),
					DiskID:   diskID,
					Path:     "/test/",
					Filename: "data.csv",
					Meta: map[string]interface{}{
						model.ArtifactInfoKey: map[string]interface{}{
							"path":     "/test/",
							"filename": "data.csv",
							"mime":     "text/csv",
							"size":     1024,
						},
					},
					AssetMeta: datatypes.NewJSONType(model.Asset{
						Bucket: "test-bucket",
						S3Key:  "test-key",
						ETag:   "test-etag",
						SHA256: "test-sha256",
						MIME:   "text/csv",
						SizeB:  1024,
					}),
				}
				m.On("GetByPath", mock.Anything, diskID, "/test/", "data.csv").Return(expectedFile, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "invalid disk ID",
			diskID:        "invalid-uuid",
			filePath:      "/test/data.csv",
			withContent:   true,
			withPublicURL: true,
			mockSetup: func(m *MockArtifactService, diskIDStr string, filePath string) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockArtifactService)
			tt.mockSetup(mockService, tt.diskID, tt.filePath)

			handler := NewArtifactHandler(mockService)

			// Create request with query parameters
			url := fmt.Sprintf("/disk/%s/artifact?file_path=%s", tt.diskID, tt.filePath)
			if tt.withContent {
				url += "&with_content=true"
			} else {
				url += "&with_content=false"
			}
			if tt.withPublicURL {
				url += "&with_public_url=true"
			} else {
				url += "&with_public_url=false"
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "disk_id", Value: tt.diskID},
			}

			// Call handler
			handler.GetArtifact(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response serializer.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)

				// Check if content is included when requested
				if tt.withContent {
					// Parse the response data to check content field
					dataBytes, _ := json.Marshal(response.Data)
					var respData map[string]interface{}
					json.Unmarshal(dataBytes, &respData)
					assert.Contains(t, respData, "content")
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
