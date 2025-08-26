package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"gorm.io/datatypes"
)

type ProjectHandler struct {
	svc service.ProjectService
}

func NewProjectHandler(s service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: s}
}

type CreateProjectReq struct {
	Configs map[string]interface{} `form:"configs" json:"configs"`
}

// CreateProject godoc
//
//	@Summary		Create project
//	@Description	Create a new project
//	@Tags			project
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		handler.CreateProjectReq	true	"CreateProject payload"
//	@Success		201		{object}	serializer.Response{data=model.Project}
//	@Router			/project [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	req := CreateProjectReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	project := model.Project{
		Configs: datatypes.JSONMap(req.Configs),
	}
	if err := h.svc.Create(c.Request.Context(), &project); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: project})
}
