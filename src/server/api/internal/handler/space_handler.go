package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/model"
	"github.com/memodb-io/Acontext/internal/serializer"
	"github.com/memodb-io/Acontext/internal/service"
	"gorm.io/datatypes"
)

type SpaceHandler struct {
	svc service.SpaceService
}

func NewSpaceHandler(s service.SpaceService) *SpaceHandler {
	return &SpaceHandler{svc: s}
}

type CreateSpaceReq struct {
	ProjectID string `form:"project_id" json:"project_id" binding:"required,uuid" format:"uuid" example:"de305d54-75b4-431b-adb2-eb6b9e546014"`
}

// CreateSpace godoc
// @Summary      Create space
// @Description  Create a new space under a project
// @Tags         space
// @Accept       json
// @Produce      json
// @Param        payload  body      handler.CreateSpaceReq  true  "CreateSpace payload"
// @Success      201      {object}  serializer.Response{data=model.Space}
// @Router       /space [post]
func (h *SpaceHandler) CreateSpace(c *gin.Context) {
	req := CreateSpaceReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	space := model.Space{
		ProjectID: datatypes.UUID(datatypes.BinUUIDFromString(req.ProjectID)),
	}
	if err := h.svc.Create(c.Request.Context(), &space); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: space})
}

// DeleteSpace godoc
// @Summary      Delete space
// @Description  Delete a space by its ID
// @Tags         space
// @Accept       json
// @Produce      json
// @Param        space_id   path      string  true  "Space ID" format(uuid)
// @Success      200        {object}  serializer.Response
// @Router       /space/{space_id} [delete]
func (h *SpaceHandler) DeleteSpace(c *gin.Context) {
	spaceID := c.Param("space_id")
	if err := h.svc.Delete(c.Request.Context(), &model.Space{ID: datatypes.UUID(datatypes.BinUUIDFromString(spaceID))}); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}
