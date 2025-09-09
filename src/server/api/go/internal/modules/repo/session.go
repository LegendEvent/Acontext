package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SessionRepo interface {
	Create(ctx context.Context, s *model.Session) error
	Delete(ctx context.Context, s *model.Session) error
	Update(ctx context.Context, s *model.Session) error
	Get(ctx context.Context, s *model.Session) (*model.Session, error)
	CreateMessageWithAssets(ctx context.Context, msg *model.Message, assetMap map[int]*model.Asset) error
	ListBySessionWithCursor(ctx context.Context, sessionID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]model.Message, error)
}

type sessionRepo struct{ db *gorm.DB }

func NewSessionRepo(db *gorm.DB) SessionRepo {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sessionRepo) Delete(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Delete(s).Error
}

func (r *sessionRepo) Update(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Where(&model.Session{ID: s.ID}).Updates(s).Error
}

func (r *sessionRepo) Get(ctx context.Context, s *model.Session) (*model.Session, error) {
	return s, r.db.WithContext(ctx).Where(&model.Session{ID: s.ID}).First(s).Error
}

func (r *sessionRepo) CreateMessageWithAssets(ctx context.Context, msg *model.Message, assetMap map[int]*model.Asset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First get the message parent id in session
		parent := model.Message{}
		if err := tx.Where(&model.Message{SessionID: msg.SessionID}).Order("created_at desc").Limit(1).Find(&parent).Error; err == nil {
			if parent.ID != uuid.Nil {
				msg.ParentID = &parent.ID
			}
		}

		// 1) upsert assets (by unique key bucket + s3_key)
		for partIdx, a := range assetMap {
			if a.Bucket == "" || a.S3Key == "" {
				return errors.New("asset missing bucket or s3_key")
			}

			// INSERT ... ON CONFLICT(bucket, s3_key) DO UPDATE SET ... RETURNING *
			if err := tx.
				Clauses(
					clause.OnConflict{
						Columns: []clause.Column{{Name: "bucket"}, {Name: "s3_key"}},
						DoUpdates: clause.Assignments(map[string]interface{}{
							"etag":        a.ETag,
							"sha256":      a.SHA256,
							"mime":        a.MIME,
							"size_bigint": a.SizeB,
						}),
					},
				).
				Create(a).Error; err != nil {
				return err
			}

			msg.Parts.Data()[partIdx].AssetID = &a.ID
		}

		// 2) Create message
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		// 3) Establish message_assets association (to avoid duplication)
		if len(assetMap) > 0 {
			links := make([]model.MessageAsset, 0, len(assetMap))
			for _, a := range assetMap {
				links = append(links, model.MessageAsset{
					MessageID: msg.ID,
					AssetID:   a.ID,
				})
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error; err != nil {
				return err
			}

			// preload message assets
			if err := tx.Preload("Assets").Where(&model.Message{ID: msg.ID}).First(msg).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *sessionRepo) ListBySessionWithCursor(ctx context.Context, sessionID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]model.Message, error) {
	q := r.db.WithContext(ctx).Where("session_id = ?", sessionID)

	// Use the (created_at, id) composite cursor; an empty cursor indicates starting from "latest"
	if !afterCreatedAt.IsZero() && afterID != uuid.Nil {
		// Retrieve strictly "older" records (reverse pagination)
		// (created_at, id) < (afterCreatedAt, afterID)
		q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", afterCreatedAt, afterCreatedAt, afterID)
	}

	var items []model.Message
	err := q.
		Preload("Assets").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&items).Error

	return items, err
}
