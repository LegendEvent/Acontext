package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/infra/blob"
	mq "github.com/memodb-io/Acontext/internal/infra/queue"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
	"github.com/memodb-io/Acontext/internal/pkg/types"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type SessionService interface {
	Create(ctx context.Context, ss *model.Session) error
	Delete(ctx context.Context, projectID uuid.UUID, sessionID uuid.UUID) error
	UpdateByID(ctx context.Context, ss *model.Session) error
	GetByID(ctx context.Context, ss *model.Session) (*model.Session, error)
	SendMessage(ctx context.Context, in SendMessageInput) (*model.Message, error)
}

type sessionService struct {
	r    repo.SessionRepo
	log  *zap.Logger
	blob *blob.S3Deps
	mq   *amqp.Connection
}

func NewSessionService(r repo.SessionRepo, log *zap.Logger, blob *blob.S3Deps, mq *amqp.Connection) SessionService {
	return &sessionService{
		r:    r,
		log:  log,
		blob: blob,
		mq:   mq,
	}
}

func (s *sessionService) Create(ctx context.Context, ss *model.Session) error {
	return s.r.Create(ctx, ss)
}

func (s *sessionService) Delete(ctx context.Context, projectID uuid.UUID, sessionID uuid.UUID) error {
	if len(sessionID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Delete(ctx, &model.Session{ID: sessionID, ProjectID: projectID})
}

func (s *sessionService) UpdateByID(ctx context.Context, ss *model.Session) error {
	return s.r.Update(ctx, ss)
}

func (s *sessionService) GetByID(ctx context.Context, ss *model.Session) (*model.Session, error) {
	if len(ss.ID) == 0 {
		return nil, errors.New("space id is empty")
	}
	return s.r.Get(ctx, ss)
}

type SendMessageInput struct {
	SessionID uuid.UUID
	Role      string
	Parts     []types.PartIn
	Files     map[string]*multipart.FileHeader
}

func (s *sessionService) SendMessage(ctx context.Context, in SendMessageInput) (*model.Message, error) {
	parts := make([]model.Part, 0, len(in.Parts))
	assets := make([]*model.Asset, 0)

	for idx, p := range in.Parts {
		tp := strings.ToLower(p.Type)
		switch tp {
		case "text", "data", "tool-call", "tool-result":
			parts = append(parts, model.Part{
				Type:     p.Type,
				Text:     p.Text,
				Markdown: false,
				Lang:     "",
				Meta:     p.Meta,
			})

		default:
			if p.FileField == "" {
				return nil, fmt.Errorf("parts[%d]: file_field required for type=%s", idx, p.Type)
			}
			fh, ok := in.Files[p.FileField]
			if !ok || fh == nil {
				return nil, fmt.Errorf("parts[%d]: missing uploaded file %s", idx, p.FileField)
			}

			// 上传到 S3
			umeta, err := s.blob.UploadFormFile(ctx, fh)
			if err != nil {
				return nil, fmt.Errorf("upload %s failed: %w", p.FileField, err)
			}

			a := &model.Asset{
				ID:       uuid.New(),
				Bucket:   umeta.Bucket,
				S3Key:    umeta.Key,
				ETag:     umeta.ETag,
				SHA256:   umeta.SHA256,
				MIME:     umeta.MIME,
				SizeB:    umeta.SizeB,
				Width:    umeta.Width,
				Height:   umeta.Height,
				Duration: umeta.Duration,
			}
			assets = append(assets, a)

			parts = append(parts, model.Part{
				AssetID:   &a.ID,
				Type:      p.Type,
				MIME:      umeta.MIME,
				SizeB:     &umeta.SizeB,
				Width:     umeta.Width,
				Height:    umeta.Height,
				DurationS: umeta.Duration,
				Meta:      p.Meta,
				Filename:  fh.Filename,
			})

		}

	}

	msg := model.Message{
		SessionID: in.SessionID,
		Role:      in.Role,
		Parts:     datatypes.NewJSONType(parts),
	}

	if err := s.r.CreateMessageWithAssets(ctx, &msg, assets); err != nil {
		return nil, err
	}

	if s.mq != nil {
		p, err := mq.NewPublisher(s.mq, "session_message", s.log)
		if err != nil {
			return nil, fmt.Errorf("create session message publisher: %w", err)
		}
		if err := p.PublishJSON(ctx, msg); err != nil {
			return nil, fmt.Errorf("publish session message: %w", err)
		}
	}

	return &msg, nil
}
