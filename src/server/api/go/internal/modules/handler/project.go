package handler

import (
	"github.com/memodb-io/Acontext/internal/modules/service"
)

type ProjectHandler struct {
	svc service.ProjectService
}

func NewProjectHandler(s service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: s}
}
