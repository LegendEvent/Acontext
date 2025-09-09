package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/config"
	"github.com/memodb-io/Acontext/internal/infra/blob"
	mq "github.com/memodb-io/Acontext/internal/infra/queue"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
	"github.com/memodb-io/Acontext/internal/pkg/paging"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type SessionService interface {
	Create(ctx context.Context, ss *model.Session) error
	Delete(ctx context.Context, projectID uuid.UUID, sessionID uuid.UUID) error
	UpdateByID(ctx context.Context, ss *model.Session) error
	GetByID(ctx context.Context, ss *model.Session) (*model.Session, error)
	SendMessage(ctx context.Context, in SendMessageInput) (*model.Message, error)
	GetMessages(ctx context.Context, in GetMessagesInput) (*GetMessagesOutput, error)
}

type sessionService struct {
	r         repo.SessionRepo
	log       *zap.Logger
	blob      *blob.S3Deps
	publisher *mq.Publisher
	cfg       *config.Config
}

func NewSessionService(r repo.SessionRepo, log *zap.Logger, blob *blob.S3Deps, publisher *mq.Publisher, cfg *config.Config) SessionService {
	return &sessionService{
		r:         r,
		log:       log,
		blob:      blob,
		publisher: publisher,
		cfg:       cfg,
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
	ProjectID uuid.UUID
	SessionID uuid.UUID
	Role      string
	Parts     []PartIn
	Files     map[string]*multipart.FileHeader
}

type SendMQPublishJSON struct {
	ProjectID uuid.UUID `json:"project_id"`
	SessionID uuid.UUID `json:"session_id"`
	MessageID uuid.UUID `json:"message_id"`
}

type PartIn struct {
	Type      string                 `json:"type"`                 // "text" | "image" | ...
	Text      string                 `json:"text,omitempty"`       // Text sharding
	FileField string                 `json:"file_field,omitempty"` // File field name in the form
	Meta      map[string]interface{} `json:"meta,omitempty"`       // [Optional] metadata
}

func (s *sessionService) SendMessage(ctx context.Context, in SendMessageInput) (*model.Message, error) {
	parts := make([]model.Part, 0, len(in.Parts))
	assetMap := make(map[int]*model.Asset)

	for idx, p := range in.Parts {
		part := model.Part{
			Type: p.Type,
			Meta: p.Meta,
		}

		if p.FileField != "" {
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
				Bucket: umeta.Bucket,
				S3Key:  umeta.Key,
				ETag:   umeta.ETag,
				SHA256: umeta.SHA256,
				MIME:   umeta.MIME,
				SizeB:  umeta.SizeB,
			}

			assetMap[idx] = a

			part.Filename = fh.Filename
			part.MIME = umeta.MIME
			part.SizeB = &umeta.SizeB
		}

		if p.Text != "" {
			part.Text = p.Text
		}

		parts = append(parts, part)

	}

	msg := model.Message{
		SessionID: in.SessionID,
		Role:      in.Role,
		Parts:     datatypes.NewJSONType(parts),
	}

	if err := s.r.CreateMessageWithAssets(ctx, &msg, assetMap); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishJSON(ctx, s.cfg.RabbitMQ.ExchangeName.SessionMessage, s.cfg.RabbitMQ.RoutingKey.SessionMessageInsert, SendMQPublishJSON{
			ProjectID: in.ProjectID,
			SessionID: in.SessionID,
			MessageID: msg.ID,
		}); err != nil {
			return nil, fmt.Errorf("publish session message: %w", err)
		}
	}

	return &msg, nil
}

type GetMessagesInput struct {
	SessionID          uuid.UUID     `json:"session_id"`
	Limit              int           `json:"limit"`
	Cursor             string        `json:"cursor"`
	WithAssetPublicURL bool          `json:"with_public_url"`
	AssetExpire        time.Duration `json:"asset_expire"`
}

type PublicURL struct {
	URL      string    `json:"url"`
	ExpireAt time.Time `json:"expire_at"`
}

type GetMessagesOutput struct {
	Items      []model.Message         `json:"items"`
	NextCursor string                  `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
	PublicURLs map[uuid.UUID]PublicURL `json:"public_urls,omitempty"` // asset_id -> url
}

func (s *sessionService) GetMessages(ctx context.Context, in GetMessagesInput) (*GetMessagesOutput, error) {
	// Parse cursor (createdAt, id); an empty cursor indicates starting from the latest
	var afterT time.Time
	var afterID uuid.UUID
	var err error
	if in.Cursor != "" {
		afterT, afterID, err = paging.DecodeCursor(in.Cursor)
		if err != nil {
			return nil, err
		}
	}

	// Query limit+1 is used to determine has_more
	msgs, err := s.r.ListBySessionWithCursor(ctx, in.SessionID, afterT, afterID, in.Limit+1)
	if err != nil {
		return nil, err
	}

	out := &GetMessagesOutput{
		Items:   msgs,
		HasMore: false,
	}
	if len(msgs) > in.Limit {
		out.HasMore = true
		out.Items = msgs[:in.Limit]
		last := out.Items[len(out.Items)-1]
		out.NextCursor = paging.EncodeCursor(last.CreatedAt, last.ID)
	}

	if in.WithAssetPublicURL {
		out.PublicURLs = make(map[uuid.UUID]PublicURL)
		for _, m := range out.Items {
			for _, a := range m.Assets {
				url, err := s.blob.PresignGet(ctx, a.S3Key, in.AssetExpire)
				if err != nil {
					return nil, fmt.Errorf("get presigned url for asset %s: %w", a.ID, err)
				}
				out.PublicURLs[a.ID] = PublicURL{
					URL:      url,
					ExpireAt: time.Now().Add(in.AssetExpire),
				}
			}
		}
	}

	return out, nil
}
