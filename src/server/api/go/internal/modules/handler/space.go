package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"gorm.io/datatypes"
)

type SpaceHandler struct {
	svc service.SpaceService
}

func NewSpaceHandler(s service.SpaceService) *SpaceHandler {
	return &SpaceHandler{svc: s}
}

type CreateSpaceReq struct {
	ProjectID string `form:"project_id" json:"project_id" binding:"required,uuid" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// CreateSpace godoc
//
//	@Summary		Create space
//	@Description	Create a new space under a project
//	@Tags			space
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		handler.CreateSpaceReq	true	"CreateSpace payload"
//	@Success		201		{object}	serializer.Response{data=model.Space}
//	@Router			/space [post]
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
//
//	@Summary		Delete space
//	@Description	Delete a space by its ID
//	@Tags			space
//	@Accept			json
//	@Produce		json
//	@Param			space_id	path		string	true	"Space ID"	Format(uuid)	Example(123e4567-e89b-12d3-a456-426614174000)
//	@Success		200			{object}	serializer.Response
//	@Router			/space/{space_id} [delete]
func (h *SpaceHandler) DeleteSpace(c *gin.Context) {
	spaceID := c.Param("space_id")
	if err := h.svc.Delete(c.Request.Context(), &model.Space{ID: datatypes.UUID(datatypes.BinUUIDFromString(spaceID))}); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}

type UpdateConfigsReq struct {
	Configs map[string]interface{} `form:"configs" json:"configs" binding:"required"`
}

// UpdateConfigs godoc
//
//	@Summary		Update space configs
//	@Description	Update the configurations of a space by its ID
//	@Tags			space
//	@Accept			json
//	@Produce		json
//	@Param			space_id	path		string						true	"Space ID"	Format(uuid)	Example(123e4567-e89b-12d3-a456-426614174000)
//	@Param			payload		body		handler.UpdateConfigsReq	true	"UpdateConfigs payload"
//	@Success		200			{object}	serializer.Response
//	@Router			/space/{space_id}/configs [put]
func (h *SpaceHandler) UpdateConfigs(c *gin.Context) {
	spaceID := c.Param("space_id")
	req := UpdateConfigsReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}
	if err := h.svc.UpdateConfigs(c.Request.Context(), &model.Space{
		ID:      datatypes.UUID(datatypes.BinUUIDFromString(spaceID)),
		Configs: datatypes.JSONMap(req.Configs),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}
