package repo

import (
	"context"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
)

type ProjectRepo interface {
	Create(ctx context.Context, p *model.Project) error
}

type projectRepo struct{ db *gorm.DB }

func NewProjectRepo(db *gorm.DB) ProjectRepo {
	return &projectRepo{db: db}
}

func (r *projectRepo) Create(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Create(p).Error
}
