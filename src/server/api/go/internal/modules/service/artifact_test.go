package service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/infra/blob"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/pkg/utils/fileparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// MockArtifactRepo is a mock implementation of ArtifactRepo
type MockArtifactRepo struct {
	mock.Mock
}

func (m *MockArtifactRepo) Create(ctx context.Context, f *model.Artifact) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockArtifactRepo) Delete(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) error {
	args := m.Called(ctx, diskID, artifactID)
	return args.Error(0)
}

func (m *MockArtifactRepo) DeleteByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) error {
	args := m.Called(ctx, diskID, path, filename)
	return args.Error(0)
}

func (m *MockArtifactRepo) Update(ctx context.Context, f *model.Artifact) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockArtifactRepo) GetByID(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, artifactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactRepo) GetByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) (*model.Artifact, error) {
	args := m.Called(ctx, diskID, path, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Artifact), args.Error(1)
}

func (m *MockArtifactRepo) ListByPath(ctx context.Context, diskID uuid.UUID, path string) ([]*model.Artifact, error) {
	args := m.Called(ctx, diskID, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Artifact), args.Error(1)
}

func (m *MockArtifactRepo) GetAllPaths(ctx context.Context, diskID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, diskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockArtifactRepo) ExistsByPathAndFilename(ctx context.Context, diskID uuid.UUID, path string, filename string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, diskID, path, filename, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockArtifactRepo) GetByDiskID(ctx context.Context, diskID uuid.UUID) ([]*model.Artifact, error) {
	args := m.Called(ctx, diskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Artifact), args.Error(1)
}

// MockArtifactS3Deps is a mock implementation of blob.S3Deps for file service
type MockArtifactS3Deps struct {
	mock.Mock
}

func (m *MockArtifactS3Deps) UploadFormFile(ctx context.Context, s3Key string, fileHeader *multipart.FileHeader) (*blob.UploadedMeta, error) {
	args := m.Called(ctx, s3Key, fileHeader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blob.UploadedMeta), args.Error(1)
}

func (m *MockArtifactS3Deps) PresignGet(ctx context.Context, s3Key string, expire time.Duration) (string, error) {
	args := m.Called(ctx, s3Key, expire)
	return args.String(0), args.Error(1)
}

// Helper functions for creating test data
func createTestArtifact() *model.Artifact {
	diskID := uuid.New()
	artifactID := uuid.New()

	return &model.Artifact{
		ID:       artifactID,
		DiskID:   diskID,
		Path:     "/test/path",
		Filename: "test.txt",
		Meta: map[string]interface{}{
			model.ArtifactInfoKey: map[string]interface{}{
				"path":     "/test/path",
				"filename": "test.txt",
				"mime":     "text/plain",
				"size":     int64(1024),
			},
		},
		AssetMeta: datatypes.NewJSONType(model.Asset{
			Bucket: "test-bucket",
			S3Key:  "artifacts/" + diskID.String() + "/test.txt",
			ETag:   "test-etag",
			SHA256: "test-sha256",
			MIME:   "text/plain",
			SizeB:  1024,
		}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestArtifactHeader() *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: "test.txt",
		Size:     1024,
	}
}

func createTestUploadedMeta() *blob.UploadedMeta {
	return &blob.UploadedMeta{
		Bucket: "test-bucket",
		Key:    "artifacts/test-artifact/test.txt",
		ETag:   "test-etag",
		SHA256: "test-sha256",
		MIME:   "text/plain",
		SizeB:  1024,
	}
}

// testArtifactService is a test version that uses interfaces
type testArtifactService struct {
	r  *MockArtifactRepo
	s3 *MockArtifactS3Deps
}

func newTestArtifactService(r *MockArtifactRepo, s3 *MockArtifactS3Deps) ArtifactService {
	return &testArtifactService{r: r, s3: s3}
}

func (s *testArtifactService) Create(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, userMeta map[string]interface{}) (*model.Artifact, error) {
	// Check if file with same path and filename already exists in the same artifact
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, path, filename, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("artifact already exists")
	}

	// Generate S3 key
	s3Key := "artifacts/" + diskID.String() + "/" + filename

	uploadedMeta, err := s.s3.UploadFormFile(ctx, s3Key, fileHeader)
	if err != nil {
		return nil, err
	}

	fileMeta := NewArtifactMetadataFromUpload(path, fileHeader, uploadedMeta)

	// Create file record with separated metadata
	meta := map[string]interface{}{
		model.ArtifactInfoKey: fileMeta.ToSystemMeta(),
	}

	for k, v := range userMeta {
		meta[k] = v
	}

	file := &model.Artifact{
		ID:        uuid.New(),
		DiskID:    diskID,
		Path:      path,
		Filename:  filename,
		Meta:      meta,
		AssetMeta: datatypes.NewJSONType(fileMeta.ToAsset()),
	}

	if err := s.r.Create(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *testArtifactService) Delete(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) error {
	if artifactID == uuid.Nil {
		return errors.New("artifact id is empty")
	}
	return s.r.Delete(ctx, diskID, artifactID)
}

func (s *testArtifactService) DeleteByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) error {
	if path == "" || filename == "" {
		return errors.New("path and filename are required")
	}
	return s.r.DeleteByPath(ctx, diskID, path, filename)
}

func (s *testArtifactService) GetByID(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) (*model.Artifact, error) {
	if artifactID == uuid.Nil {
		return nil, errors.New("artifact id is empty")
	}
	return s.r.GetByID(ctx, diskID, artifactID)
}

func (s *testArtifactService) GetByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) (*model.Artifact, error) {
	if path == "" || filename == "" {
		return nil, errors.New("path and filename are required")
	}
	return s.r.GetByPath(ctx, diskID, path, filename)
}

func (s *testArtifactService) GetPresignedURL(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, expire time.Duration) (string, error) {
	file, err := s.GetByID(ctx, diskID, artifactID)
	if err != nil {
		return "", err
	}

	assetData := file.AssetMeta.Data()
	if assetData.S3Key == "" {
		return "", errors.New("artifact has no S3 key")
	}

	return s.s3.PresignGet(ctx, assetData.S3Key, expire)
}

func (s *testArtifactService) GetPresignedURLByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, expire time.Duration) (string, error) {
	file, err := s.GetByPath(ctx, diskID, path, filename)
	if err != nil {
		return "", err
	}

	assetData := file.AssetMeta.Data()
	if assetData.S3Key == "" {
		return "", errors.New("artifact has no S3 key")
	}

	return s.s3.PresignGet(ctx, assetData.S3Key, expire)
}

func (s *testArtifactService) UpdateArtifact(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	// Get existing file
	file, err := s.GetByID(ctx, diskID, artifactID)
	if err != nil {
		return nil, err
	}

	// Determine the target path and filename
	var path, filename string
	if newPath != nil && *newPath != "" {
		path = *newPath
	} else {
		path = file.Path
	}

	if newFilename != nil && *newFilename != "" {
		filename = *newFilename
	} else {
		filename = file.Filename
	}

	// Check if file with same path and filename already exists for another file in the same artifact
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, path, filename, &artifactID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("artifact already exists")
	}

	// Generate new S3 key
	s3Key := "artifacts/" + diskID.String() + "/" + filename

	uploadedMeta, err := s.s3.UploadFormFile(ctx, s3Key, fileHeader)
	if err != nil {
		return nil, err
	}

	fileMeta := NewArtifactMetadataFromUpload(path, fileHeader, uploadedMeta)

	// Update file record
	file.Path = path
	file.Filename = filename
	file.AssetMeta = datatypes.NewJSONType(fileMeta.ToAsset())

	// Update system meta with new file info
	systemMeta, ok := file.Meta[model.ArtifactInfoKey].(map[string]interface{})
	if !ok {
		systemMeta = make(map[string]interface{})
		file.Meta[model.ArtifactInfoKey] = systemMeta
	}

	// Update system metadata
	for k, v := range fileMeta.ToSystemMeta() {
		systemMeta[k] = v
	}

	if err := s.r.Update(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *testArtifactService) UpdateArtifactByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	// Get existing file
	file, err := s.GetByPath(ctx, diskID, path, filename)
	if err != nil {
		return nil, err
	}

	// Determine the target path and filename
	var targetPath, targetFilename string
	if newPath != nil && *newPath != "" {
		targetPath = *newPath
	} else {
		targetPath = file.Path
	}

	if newFilename != nil && *newFilename != "" {
		targetFilename = *newFilename
	} else {
		targetFilename = file.Filename
	}

	// Check if file with same path and filename already exists for another file in the same artifact
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, targetPath, targetFilename, &file.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("artifact already exists")
	}

	// Generate new S3 key
	s3Key := "artifacts/" + diskID.String() + "/" + targetFilename

	uploadedMeta, err := s.s3.UploadFormFile(ctx, s3Key, fileHeader)
	if err != nil {
		return nil, err
	}

	fileMeta := NewArtifactMetadataFromUpload(targetPath, fileHeader, uploadedMeta)

	// Update file record
	file.Path = targetPath
	file.Filename = targetFilename
	file.AssetMeta = datatypes.NewJSONType(fileMeta.ToAsset())

	// Update system meta with new file info
	systemMeta, ok := file.Meta[model.ArtifactInfoKey].(map[string]interface{})
	if !ok {
		systemMeta = make(map[string]interface{})
		file.Meta[model.ArtifactInfoKey] = systemMeta
	}

	// Update system metadata
	for k, v := range fileMeta.ToSystemMeta() {
		systemMeta[k] = v
	}

	if err := s.r.Update(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *testArtifactService) ListByPath(ctx context.Context, diskID uuid.UUID, path string) ([]*model.Artifact, error) {
	return s.r.ListByPath(ctx, diskID, path)
}

func (s *testArtifactService) GetAllPaths(ctx context.Context, diskID uuid.UUID) ([]string, error) {
	return s.r.GetAllPaths(ctx, diskID)
}

func (s *testArtifactService) GetByDiskID(ctx context.Context, diskID uuid.UUID) ([]*model.Artifact, error) {
	return s.r.GetByDiskID(ctx, diskID)
}

func (s *testArtifactService) GetFileContent(ctx context.Context, diskID uuid.UUID, path string, filename string) (*fileparser.FileContent, error) {
	// This is a test implementation that doesn't actually download from S3
	// In real tests, you would mock the S3 download and file parsing
	return &fileparser.FileContent{
		Type: "text",
		Raw:  "test content",
	}, nil
}

// Test cases for Create method
func TestArtifactService_Create(t *testing.T) {
	diskID := uuid.New()
	path := "/test/path"
	filename := "test.txt"
	fileHeader := createTestArtifactHeader()
	userMeta := map[string]interface{}{
		"custom_key": "custom_value",
	}

	tests := []struct {
		name        string
		setup       func(*MockArtifactRepo, *MockArtifactS3Deps)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation",
			setup: func(repo *MockArtifactRepo, s3 *MockArtifactS3Deps) {
				repo.On("ExistsByPathAndFilename", mock.Anything, diskID, path, filename, (*uuid.UUID)(nil)).Return(false, nil)
				s3.On("UploadFormFile", mock.Anything, mock.AnythingOfType("string"), fileHeader).Return(createTestUploadedMeta(), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(f *model.Artifact) bool {
					return f.DiskID == diskID && f.Path == path && f.Filename == filename
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name: "artifact already exists",
			setup: func(repo *MockArtifactRepo, s3 *MockArtifactS3Deps) {
				repo.On("ExistsByPathAndFilename", mock.Anything, diskID, path, filename, (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectError: true,
			errorMsg:    "artifact already exists",
		},
		{
			name: "upload error",
			setup: func(repo *MockArtifactRepo, s3 *MockArtifactS3Deps) {
				repo.On("ExistsByPathAndFilename", mock.Anything, diskID, path, filename, (*uuid.UUID)(nil)).Return(false, nil)
				s3.On("UploadFormFile", mock.Anything, mock.AnythingOfType("string"), fileHeader).Return(nil, errors.New("upload error"))
			},
			expectError: true,
			errorMsg:    "upload error",
		},
		{
			name: "create record error",
			setup: func(repo *MockArtifactRepo, s3 *MockArtifactS3Deps) {
				repo.On("ExistsByPathAndFilename", mock.Anything, diskID, path, filename, (*uuid.UUID)(nil)).Return(false, nil)
				s3.On("UploadFormFile", mock.Anything, mock.AnythingOfType("string"), fileHeader).Return(createTestUploadedMeta(), nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("create error"))
			},
			expectError: true,
			errorMsg:    "create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockArtifactRepo{}
			mockS3 := &MockArtifactS3Deps{}
			tt.setup(mockRepo, mockS3)

			service := newTestArtifactService(mockRepo, mockS3)

			file, err := service.Create(context.Background(), diskID, path, filename, fileHeader, userMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, file)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
				assert.Equal(t, diskID, file.DiskID)
				assert.Equal(t, path, file.Path)
				assert.Equal(t, filename, file.Filename)
				assert.Contains(t, file.Meta, model.ArtifactInfoKey)
				assert.Contains(t, file.Meta, "custom_key")
			}

			mockRepo.AssertExpectations(t)
			mockS3.AssertExpectations(t)
		})
	}
}

// Test cases for Delete method
func TestArtifactService_Delete(t *testing.T) {
	diskID := uuid.New()
	artifactID := uuid.New()

	tests := []struct {
		name        string
		artifactID  uuid.UUID
		setup       func(*MockArtifactRepo)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "successful deletion",
			artifactID: artifactID,
			setup: func(repo *MockArtifactRepo) {
				repo.On("Delete", mock.Anything, diskID, artifactID).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "empty file ID",
			artifactID:  uuid.UUID{},
			setup:       func(repo *MockArtifactRepo) {},
			expectError: true,
			errorMsg:    "artifact id is empty",
		},
		{
			name:       "repo error",
			artifactID: artifactID,
			setup: func(repo *MockArtifactRepo) {
				repo.On("Delete", mock.Anything, diskID, artifactID).Return(errors.New("delete error"))
			},
			expectError: true,
			errorMsg:    "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockArtifactRepo{}
			tt.setup(mockRepo)

			service := newTestArtifactService(mockRepo, &MockArtifactS3Deps{})

			err := service.Delete(context.Background(), diskID, tt.artifactID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			if tt.errorMsg != "artifact id is empty" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// Test cases for GetByID method
func TestArtifactService_GetByID(t *testing.T) {
	diskID := uuid.New()
	artifactID := uuid.New()
	testFile := createTestArtifact()
	testFile.ID = artifactID
	testFile.DiskID = diskID

	tests := []struct {
		name        string
		artifactID  uuid.UUID
		setup       func(*MockArtifactRepo)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "successful retrieval",
			artifactID: artifactID,
			setup: func(repo *MockArtifactRepo) {
				repo.On("GetByID", mock.Anything, diskID, artifactID).Return(testFile, nil)
			},
			expectError: false,
		},
		{
			name:        "empty file ID",
			artifactID:  uuid.UUID{},
			setup:       func(repo *MockArtifactRepo) {},
			expectError: true,
			errorMsg:    "artifact id is empty",
		},
		{
			name:       "file not found",
			artifactID: artifactID,
			setup: func(repo *MockArtifactRepo) {
				repo.On("GetByID", mock.Anything, diskID, artifactID).Return(nil, errors.New("file not found"))
			},
			expectError: true,
			errorMsg:    "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockArtifactRepo{}
			tt.setup(mockRepo)

			service := newTestArtifactService(mockRepo, &MockArtifactS3Deps{})

			file, err := service.GetByID(context.Background(), diskID, tt.artifactID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, file)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
				assert.Equal(t, artifactID, file.ID)
				assert.Equal(t, diskID, file.DiskID)
			}

			if tt.errorMsg != "artifact id is empty" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
