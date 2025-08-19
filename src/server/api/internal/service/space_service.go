package service

import (
	"context"

	"github.com/memodb-io/Acontext/internal/model"
	"github.com/memodb-io/Acontext/internal/repo"
)

type SpaceService interface {
	Create(ctx context.Context, m *model.Space) error
	Delete(ctx context.Context, m *model.Space) error
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
