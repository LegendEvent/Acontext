package service

import (
	"context"
	"errors"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
	"gorm.io/datatypes"
)

type SpaceService interface {
	Create(ctx context.Context, m *model.Space) error
	Delete(ctx context.Context, spaceID string) error
	UpdateByID(ctx context.Context, m *model.Space) error
	GetByID(ctx context.Context, m *model.Space) (*model.Space, error)
}

type spaceService struct{ r repo.SpaceRepo }

func NewSpaceService(r repo.SpaceRepo) SpaceService {
	return &spaceService{r: r}
}

func (s *spaceService) Create(ctx context.Context, m *model.Space) error {
	return s.r.Create(ctx, m)
}

func (s *spaceService) Delete(ctx context.Context, spaceID string) error {
	if len(spaceID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Delete(ctx, &model.Space{ID: datatypes.UUID(datatypes.BinUUIDFromString(spaceID))})
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
