package repo

import (
	"context"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
)

type SessionRepo interface {
	Create(ctx context.Context, s *model.Session) error
}

type sessionRepo struct{ db *gorm.DB }

func NewSessionRepo(db *gorm.DB) SessionRepo {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}
