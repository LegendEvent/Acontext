package service

import (
	"context"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
)

type SessionService interface {
	Create(ctx context.Context, ss *model.Session) error
}

type sessionService struct{ r repo.SessionRepo }

func NewSessionService(r repo.SessionRepo) SessionService {
	return &sessionService{r: r}
}

func (s *sessionService) Create(ctx context.Context, ss *model.Session) error {
	return s.r.Create(ctx, ss)
}
