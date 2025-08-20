package handler

import (
	"github.com/memodb-io/Acontext/internal/modules/service"
)

type SessionHandler struct {
	svc service.SessionService
}

func NewSessionHandler(s service.SessionService) *SessionHandler {
	return &SessionHandler{svc: s}
}
