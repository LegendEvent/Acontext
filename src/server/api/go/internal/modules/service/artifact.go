package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/infra/blob"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
	"github.com/memodb-io/Acontext/internal/pkg/utils/fileparser"
	"gorm.io/datatypes"
)

// ArtifactMetadata centrally manages artifact-related metadata
type ArtifactMetadata struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	MIME     string `json:"mime"`
	SizeB    int64  `json:"size_b"`
	Bucket   string `json:"bucket"`
	S3Key    string `json:"s3_key"`
	ETag     string `json:"etag"`
	SHA256   string `json:"sha256"`
}

// ToAsset converts to Asset model
func (am *ArtifactMetadata) ToAsset() model.Asset {
	return model.Asset{
		Bucket: am.Bucket,
		S3Key:  am.S3Key,
		ETag:   am.ETag,
		SHA256: am.SHA256,
		MIME:   am.MIME,
		SizeB:  am.SizeB,
	}
}

// ToSystemMeta converts to system metadata
func (am *ArtifactMetadata) ToSystemMeta() map[string]interface{} {
	return map[string]interface{}{
		"path":     am.Path,
		"filename": am.Filename,
		"mime":     am.MIME,
		"size":     am.SizeB,
	}
}

// NewArtifactMetadataFromUpload creates ArtifactMetadata from the uploaded file
func NewArtifactMetadataFromUpload(path string, fileHeader *multipart.FileHeader, uploadedMeta *blob.UploadedMeta) *ArtifactMetadata {
	return &ArtifactMetadata{
		Path:     path,
		Filename: fileHeader.Filename,
		MIME:     uploadedMeta.MIME,
		SizeB:    uploadedMeta.SizeB,
		Bucket:   uploadedMeta.Bucket,
		S3Key:    uploadedMeta.Key,
		ETag:     uploadedMeta.ETag,
		SHA256:   uploadedMeta.SHA256,
	}
}

type ArtifactService interface {
	Create(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, userMeta map[string]interface{}) (*model.Artifact, error)
	Delete(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) error
	DeleteByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) error
	GetByID(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) (*model.Artifact, error)
	GetByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) (*model.Artifact, error)
	GetPresignedURL(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, expire time.Duration) (string, error)
	GetPresignedURLByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, expire time.Duration) (string, error)
	GetFileContent(ctx context.Context, diskID uuid.UUID, path string, filename string) (*fileparser.FileContent, error)
	UpdateArtifact(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error)
	UpdateArtifactByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error)
	ListByPath(ctx context.Context, diskID uuid.UUID, path string) ([]*model.Artifact, error)
	GetAllPaths(ctx context.Context, diskID uuid.UUID) ([]string, error)
	GetByDiskID(ctx context.Context, diskID uuid.UUID) ([]*model.Artifact, error)
}

type artifactService struct {
	r  repo.ArtifactRepo
	s3 *blob.S3Deps
}

func NewArtifactService(r repo.ArtifactRepo, s3 *blob.S3Deps) ArtifactService {
	return &artifactService{r: r, s3: s3}
}

func (s *artifactService) Create(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, userMeta map[string]interface{}) (*model.Artifact, error) {
	// Check if artifact with same path and filename already exists in the same disk
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, path, filename, nil)
	if err != nil {
		return nil, fmt.Errorf("check artifact existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("artifact '%s' already exists in path '%s'", filename, path)
	}

	uploadedMeta, err := s.s3.UploadFormFile(ctx, "disks/"+diskID.String(), fileHeader)
	if err != nil {
		return nil, fmt.Errorf("upload file to S3: %w", err)
	}

	artifactMeta := NewArtifactMetadataFromUpload(path, fileHeader, uploadedMeta)

	// Create artifact record with separated metadata
	meta := map[string]interface{}{
		model.ArtifactInfoKey: artifactMeta.ToSystemMeta(),
	}

	for k, v := range userMeta {
		meta[k] = v
	}

	artifact := &model.Artifact{
		ID:        uuid.New(),
		DiskID:    diskID,
		Path:      path,
		Filename:  filename,
		Meta:      meta,
		AssetMeta: datatypes.NewJSONType(artifactMeta.ToAsset()),
	}

	if err := s.r.Create(ctx, artifact); err != nil {
		return nil, fmt.Errorf("create artifact record: %w", err)
	}

	return artifact, nil
}

func (s *artifactService) Delete(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) error {
	if len(artifactID) == 0 {
		return errors.New("artifact id is empty")
	}
	return s.r.Delete(ctx, diskID, artifactID)
}

func (s *artifactService) DeleteByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) error {
	if path == "" || filename == "" {
		return errors.New("path and filename are required")
	}
	return s.r.DeleteByPath(ctx, diskID, path, filename)
}

func (s *artifactService) GetByID(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID) (*model.Artifact, error) {
	if len(artifactID) == 0 {
		return nil, errors.New("artifact id is empty")
	}
	return s.r.GetByID(ctx, diskID, artifactID)
}

func (s *artifactService) GetByPath(ctx context.Context, diskID uuid.UUID, path string, filename string) (*model.Artifact, error) {
	if path == "" || filename == "" {
		return nil, errors.New("path and filename are required")
	}
	return s.r.GetByPath(ctx, diskID, path, filename)
}

func (s *artifactService) GetPresignedURL(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, expire time.Duration) (string, error) {
	artifact, err := s.GetByID(ctx, diskID, artifactID)
	if err != nil {
		return "", err
	}

	assetData := artifact.AssetMeta.Data()
	if assetData.S3Key == "" {
		return "", errors.New("artifact has no S3 key")
	}

	return s.s3.PresignGet(ctx, assetData.S3Key, expire)
}

func (s *artifactService) GetPresignedURLByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, expire time.Duration) (string, error) {
	artifact, err := s.GetByPath(ctx, diskID, path, filename)
	if err != nil {
		return "", err
	}

	assetData := artifact.AssetMeta.Data()
	if assetData.S3Key == "" {
		return "", errors.New("artifact has no S3 key")
	}

	return s.s3.PresignGet(ctx, assetData.S3Key, expire)
}

func (s *artifactService) GetFileContent(ctx context.Context, diskID uuid.UUID, path string, filename string) (*fileparser.FileContent, error) {
	artifact, err := s.GetByPath(ctx, diskID, path, filename)
	if err != nil {
		return nil, err
	}

	assetData := artifact.AssetMeta.Data()
	if assetData.S3Key == "" {
		return nil, errors.New("artifact has no S3 key")
	}

	// Download file content from S3
	content, err := s.s3.DownloadFile(ctx, assetData.S3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download file content: %w", err)
	}

	// Parse file content
	parser := fileparser.NewFileParser()
	fileContent, err := parser.ParseFile(filename, assetData.MIME, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file content: %w", err)
	}

	return fileContent, nil
}

func (s *artifactService) UpdateArtifact(ctx context.Context, diskID uuid.UUID, artifactID uuid.UUID, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	// Get existing artifact
	artifact, err := s.GetByID(ctx, diskID, artifactID)
	if err != nil {
		return nil, err
	}

	// Determine the target path and filename
	var path, filename string
	if newPath != nil && *newPath != "" {
		path = *newPath
	} else {
		path = artifact.Path
	}

	if newFilename != nil && *newFilename != "" {
		filename = *newFilename
	} else {
		filename = artifact.Filename
	}

	// Check if artifact with same path and filename already exists for another artifact in the same disk
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, path, filename, &artifactID)
	if err != nil {
		return nil, fmt.Errorf("check artifact existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("artifact '%s' already exists in path '%s'", filename, path)
	}

	uploadedMeta, err := s.s3.UploadFormFile(ctx, "disks/"+diskID.String(), fileHeader)
	if err != nil {
		return nil, fmt.Errorf("upload file to S3: %w", err)
	}

	artifactMeta := NewArtifactMetadataFromUpload(path, fileHeader, uploadedMeta)

	// Update artifact record
	artifact.Path = path
	artifact.Filename = filename
	artifact.AssetMeta = datatypes.NewJSONType(artifactMeta.ToAsset())

	// Update system meta with new artifact info
	systemMeta, ok := artifact.Meta[model.ArtifactInfoKey].(map[string]interface{})
	if !ok {
		systemMeta = make(map[string]interface{})
		artifact.Meta[model.ArtifactInfoKey] = systemMeta
	}

	// Update system metadata
	for k, v := range artifactMeta.ToSystemMeta() {
		systemMeta[k] = v
	}

	if err := s.r.Update(ctx, artifact); err != nil {
		return nil, fmt.Errorf("update artifact record: %w", err)
	}

	return artifact, nil
}

func (s *artifactService) UpdateArtifactByPath(ctx context.Context, diskID uuid.UUID, path string, filename string, fileHeader *multipart.FileHeader, newPath *string, newFilename *string) (*model.Artifact, error) {
	// Get existing artifact
	artifact, err := s.GetByPath(ctx, diskID, path, filename)
	if err != nil {
		return nil, err
	}

	// Determine the target path and filename
	var targetPath, targetFilename string
	if newPath != nil && *newPath != "" {
		targetPath = *newPath
	} else {
		targetPath = artifact.Path
	}

	if newFilename != nil && *newFilename != "" {
		targetFilename = *newFilename
	} else {
		targetFilename = artifact.Filename
	}

	// Check if artifact with same path and filename already exists for another artifact in the same disk
	exists, err := s.r.ExistsByPathAndFilename(ctx, diskID, targetPath, targetFilename, &artifact.ID)
	if err != nil {
		return nil, fmt.Errorf("check artifact existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("artifact '%s' already exists in path '%s'", targetFilename, targetPath)
	}

	uploadedMeta, err := s.s3.UploadFormFile(ctx, "disks/"+diskID.String(), fileHeader)
	if err != nil {
		return nil, fmt.Errorf("upload file to S3: %w", err)
	}

	artifactMeta := NewArtifactMetadataFromUpload(targetPath, fileHeader, uploadedMeta)

	// Update artifact record
	artifact.Path = targetPath
	artifact.Filename = targetFilename
	artifact.AssetMeta = datatypes.NewJSONType(artifactMeta.ToAsset())

	// Update system meta with new artifact info
	systemMeta, ok := artifact.Meta[model.ArtifactInfoKey].(map[string]interface{})
	if !ok {
		systemMeta = make(map[string]interface{})
		artifact.Meta[model.ArtifactInfoKey] = systemMeta
	}

	// Update system metadata
	for k, v := range artifactMeta.ToSystemMeta() {
		systemMeta[k] = v
	}

	if err := s.r.Update(ctx, artifact); err != nil {
		return nil, fmt.Errorf("update artifact record: %w", err)
	}

	return artifact, nil
}

func (s *artifactService) ListByPath(ctx context.Context, diskID uuid.UUID, path string) ([]*model.Artifact, error) {
	return s.r.ListByPath(ctx, diskID, path)
}

func (s *artifactService) GetAllPaths(ctx context.Context, diskID uuid.UUID) ([]string, error) {
	return s.r.GetAllPaths(ctx, diskID)
}

func (s *artifactService) GetByDiskID(ctx context.Context, diskID uuid.UUID) ([]*model.Artifact, error) {
	return s.r.GetByDiskID(ctx, diskID)
}
