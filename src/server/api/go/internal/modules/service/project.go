package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
)

type ProjectService interface {
	Create(ctx context.Context, u *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateByID(ctx context.Context, u *model.Project) error
}

type projectService struct{ r repo.ProjectRepo }

func NewProjectService(r repo.ProjectRepo) ProjectService {
	return &projectService{r: r}
}

func (s *projectService) Create(ctx context.Context, u *model.Project) error {
	return s.r.Create(ctx, u)
}

func (s *projectService) Delete(ctx context.Context, id uuid.UUID) error {
	if len(id) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Delete(ctx, &model.Project{ID: id})
}

func (s *projectService) UpdateByID(ctx context.Context, u *model.Project) error {
	return s.r.Update(ctx, u)
}
