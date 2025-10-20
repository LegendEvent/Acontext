package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
)

type DiskHandler struct {
	svc service.DiskService
}

func NewDiskHandler(s service.DiskService) *DiskHandler {
	return &DiskHandler{svc: s}
}

// CreateDisk godoc
//
//	@Summary		Create disk
//	@Description	Create a disk group under a project
//	@Tags			disk
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		201	{object}	serializer.Response{data=model.Disk}
//	@Router			/disk [post]
func (h *DiskHandler) CreateDisk(c *gin.Context) {
	project, ok := c.MustGet("project").(*model.Project)
	if !ok {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", errors.New("project not found")))
		return
	}

	disk, err := h.svc.Create(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: disk})
}

// ListDisks godoc
//
//	@Summary		List disks
//	@Description	List all disks under a project
//	@Tags			disk
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{data=[]model.Disk}
//	@Router			/disk [get]
func (h *DiskHandler) ListDisks(c *gin.Context) {
	project, ok := c.MustGet("project").(*model.Project)
	if !ok {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", errors.New("project not found")))
		return
	}

	disks, err := h.svc.List(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{Data: disks})
}

// DeleteDisk godoc
//
//	@Summary		Delete disk
//	@Description	Delete a disk by its UUID
//	@Tags			disk
//	@Accept			json
//	@Produce		json
//	@Param			disk_id	path	string	true	"Disk ID"	Format(uuid)	Example(123e4567-e89b-12d3-a456-426614174000)
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{}
//	@Router			/disk/{disk_id} [delete]
func (h *DiskHandler) DeleteDisk(c *gin.Context) {
	diskID, err := uuid.Parse(c.Param("disk_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	project, ok := c.MustGet("project").(*model.Project)
	if !ok {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", errors.New("project not found")))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), project.ID, diskID); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}
