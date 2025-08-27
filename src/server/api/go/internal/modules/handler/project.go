package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/config"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/memodb-io/Acontext/internal/pkg/utils"
	"gorm.io/datatypes"
)

type ProjectHandler struct {
	svc    service.ProjectService
	Config *config.Config
}

func NewProjectHandler(s service.ProjectService, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		svc:    s,
		Config: cfg,
	}
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
//	@Param			payload	body	handler.CreateProjectReq	true	"CreateProject payload"
//	@Security		RootAuth
//	@Success		201	{object}	serializer.Response{data=model.Project}
//	@Router			/project [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	req := CreateProjectReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	key, err := utils.GenerateKey(h.Config.Root.ProjectBearerTokenPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	project := model.Project{
		SecretKey: key,
		Configs:   datatypes.JSONMap(req.Configs),
	}
	if err := h.svc.Create(c.Request.Context(), &project); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: project})
}
