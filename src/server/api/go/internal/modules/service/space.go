package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
)

type SpaceService interface {
	Create(ctx context.Context, m *model.Space) error
	Delete(ctx context.Context, projectID uuid.UUID, spaceID uuid.UUID) error
	UpdateByID(ctx context.Context, m *model.Space) error
	GetByID(ctx context.Context, m *model.Space) (*model.Space, error)
	List(ctx context.Context, projectID uuid.UUID) ([]model.Space, error)
}

type spaceService struct{ r repo.SpaceRepo }

func NewSpaceService(r repo.SpaceRepo) SpaceService {
	return &spaceService{r: r}
}

func (s *spaceService) Create(ctx context.Context, m *model.Space) error {
	return s.r.Create(ctx, m)
}

func (s *spaceService) Delete(ctx context.Context, projectID uuid.UUID, spaceID uuid.UUID) error {
	if len(spaceID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Delete(ctx, &model.Space{ID: spaceID, ProjectID: projectID})
}

func (s *spaceService) UpdateByID(ctx context.Context, m *model.Space) error {
	if len(m.ID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Update(ctx, m)
}

func (s *spaceService) GetByID(ctx context.Context, m *model.Space) (*model.Space, error) {
	if len(m.ID) == 0 {
		return nil, errors.New("space id is empty")
	}
	return s.r.Get(ctx, m)
}

func (s *spaceService) List(ctx context.Context, projectID uuid.UUID) ([]model.Space, error) {
	return s.r.List(ctx, projectID)
}
