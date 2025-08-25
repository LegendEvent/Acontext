package service

import (
	"context"
	"errors"

	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
	"gorm.io/datatypes"
)

type SessionService interface {
	Create(ctx context.Context, ss *model.Session) error
	Delete(ctx context.Context, sessionID string) error
	UpdateByID(ctx context.Context, ss *model.Session) error
	GetByID(ctx context.Context, ss *model.Session) (*model.Session, error)
}

type sessionService struct{ r repo.SessionRepo }

func NewSessionService(r repo.SessionRepo) SessionService {
	return &sessionService{r: r}
}

func (s *sessionService) Create(ctx context.Context, ss *model.Session) error {
	return s.r.Create(ctx, ss)
}

func (s *sessionService) Delete(ctx context.Context, sessionID string) error {
	if len(sessionID) == 0 {
		return errors.New("space id is empty")
	}
	return s.r.Delete(ctx, &model.Session{ID: datatypes.UUID(datatypes.BinUUIDFromString(sessionID))})
}

func (s *sessionService) UpdateByID(ctx context.Context, ss *model.Session) error {
	return s.r.Update(ctx, ss)
}

func (s *sessionService) GetByID(ctx context.Context, ss *model.Session) (*model.Session, error) {
	if len(ss.ID) == 0 {
		return nil, errors.New("space id is empty")
	}
	return s.r.Get(ctx, ss)
}
