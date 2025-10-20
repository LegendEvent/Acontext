package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
)

type DiskRepo interface {
	Create(ctx context.Context, d *model.Disk) error
	Delete(ctx context.Context, projectID uuid.UUID, diskID uuid.UUID) error
	List(ctx context.Context, projectID uuid.UUID) ([]*model.Disk, error)
}

type diskRepo struct{ db *gorm.DB }

func NewDiskRepo(db *gorm.DB) DiskRepo {
	return &diskRepo{db: db}
}

func (r *diskRepo) Create(ctx context.Context, d *model.Disk) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *diskRepo) Delete(ctx context.Context, projectID uuid.UUID, diskID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND project_id = ?", diskID, projectID).Delete(&model.Disk{}).Error
}

func (r *diskRepo) List(ctx context.Context, projectID uuid.UUID) ([]*model.Disk, error) {
	var disks []*model.Disk
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&disks).Error
	return disks, err
}
