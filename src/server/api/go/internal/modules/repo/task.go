package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"gorm.io/gorm"
)

type TaskRepo interface {
	ListBySessionWithCursor(ctx context.Context, sessionID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]model.Task, error)
}

type taskRepo struct{ db *gorm.DB }

func NewTaskRepo(db *gorm.DB) TaskRepo {
	return &taskRepo{db: db}
}

func (r *taskRepo) ListBySessionWithCursor(ctx context.Context, sessionID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]model.Task, error) {
	q := r.db.WithContext(ctx).Where("session_id = ?", sessionID)

	// Use the (created_at, id) composite cursor; an empty cursor indicates starting from "latest"
	if !afterCreatedAt.IsZero() && afterID != uuid.Nil {
		// Retrieve strictly "older" records (reverse pagination)
		// (created_at, id) < (afterCreatedAt, afterID)
		q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", afterCreatedAt, afterCreatedAt, afterID)
	}

	var items []model.Task

	return items, q.Order("created_at DESC, id DESC").Limit(limit).Find(&items).Error
}
