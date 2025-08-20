package repo

import (
	"context"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
)

type SpaceRepo interface {
	Create(ctx context.Context, s *model.Space) error
	Delete(ctx context.Context, s *model.Space) error
	Update(ctx context.Context, s *model.Space) error
}

type spaceRepo struct{ db *gorm.DB }

func NewSpaceRepo(db *gorm.DB) SpaceRepo {
	return &spaceRepo{db: db}
}

func (r *spaceRepo) Create(ctx context.Context, s *model.Space) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *spaceRepo) Delete(ctx context.Context, s *model.Space) error {
	return r.db.WithContext(ctx).Delete(s).Error
}

func (r *spaceRepo) Update(ctx context.Context, s *model.Space) error {
	return r.db.WithContext(ctx).Where(&model.Space{ID: s.ID}).Updates(s).Error
}
