package service

import (
	"context"
	"errors"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
)

type SpaceService interface {
	Create(ctx context.Context, m *model.Space) error
	Delete(ctx context.Context, m *model.Space) error
	UpdateConfigs(ctx context.Context, m *model.Space) error
}

type spaceService struct{ r repo.SpaceRepo }

func NewSpaceService(r repo.SpaceRepo) SpaceService {
	return &spaceService{r: r}
}

func (s *spaceService) Create(ctx context.Context, m *model.Space) error {
	return s.r.Create(ctx, m)
}

func (s *spaceService) Delete(ctx context.Context, m *model.Space) error {
	return s.r.Delete(ctx, m)
}

func (s *spaceService) UpdateConfigs(ctx context.Context, m *model.Space) error {
	if len(m.ID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Update(ctx, m)
}
