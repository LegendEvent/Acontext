package service

import (
	"context"

	"github.com/memodb-io/Acontext/internal/model"
	"github.com/memodb-io/Acontext/internal/repo"
)

type ProjectService interface {
	Create(ctx context.Context, u *model.Project) error
}

type projectService struct{ r repo.ProjectRepo }

func NewProjectService(r repo.ProjectRepo) ProjectService {
	return &projectService{r: r}
}

func (s *projectService) Create(ctx context.Context, u *model.Project) error {
	return s.r.Create(ctx, u)
}
