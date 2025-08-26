package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/modules/dto"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"gorm.io/datatypes"
)

type SessionHandler struct {
	svc service.SessionService
}

func NewSessionHandler(s service.SessionService) *SessionHandler {
	return &SessionHandler{svc: s}
}

type CreateSessionReq struct {
	ProjectID string                 `form:"project_id" json:"project_id" binding:"required,uuid" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	SpaceID   string                 `form:"space_id" json:"space_id" format:"uuid" example:"123e4567-e89b-12d3-a456-42661417"`
	Configs   map[string]interface{} `form:"configs" json:"configs"`
}

// CreateSession godoc
//
//	@Summary		Create session
//	@Description	Create a new session under a space
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		handler.CreateSessionReq	true	"CreateSession payload"
//	@Success		201		{object}	serializer.Response{data=model.Session}
//	@Router			/session [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	req := CreateSessionReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	session := model.Session{
		ProjectID: datatypes.UUID(datatypes.BinUUIDFromString(req.ProjectID)),
		SpaceID:   datatypes.UUID(datatypes.BinUUIDFromString(req.SpaceID)),
		Configs:   datatypes.JSONMap(req.Configs),
	}
	if err := h.svc.Create(c.Request.Context(), &session); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: session})
}

// DeleteSession godoc
//
//	@Summary		Delete session
//	@Description	Delete a session by id
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string	true	"Session ID"	format(uuid)
//	@Success		200			{object}	serializer.Response{}
//	@Router			/session/{session_id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if err := h.svc.Delete(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}

type UpdateSessionConfigsReq struct {
	Configs map[string]interface{} `form:"configs" json:"configs"`
}

// UpdateSessionConfigs godoc
//
//	@Summary		Update session configs
//	@Description	Update session configs by id
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string							true	"Session ID"	format(uuid)
//	@Param			payload		body		handler.UpdateSessionConfigsReq	true	"UpdateSessionConfigs payload"
//	@Success		200			{object}	serializer.Response{}
//	@Router			/session/{session_id}/configs [put]
func (h *SessionHandler) UpdateConfigs(c *gin.Context) {
	req := UpdateSessionConfigsReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	sessionID := c.Param("session_id")
	if err := h.svc.UpdateByID(c.Request.Context(), &model.Session{
		ID:      datatypes.UUID(datatypes.BinUUIDFromString(sessionID)),
		Configs: datatypes.JSONMap(req.Configs),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}

// GetSessionConfigs godoc
//
//	@Summary		Get session configs
//	@Description	Get session configs by id
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string	true	"Session ID"	format(uuid)
//	@Success		200			{object}	serializer.Response{data=model.Session}
//	@Router			/session/{session_id}/configs [get]
func (h *SessionHandler) GetConfigs(c *gin.Context) {
	sessionID := c.Param("session_id")
	session, err := h.svc.GetByID(c.Request.Context(), &model.Session{ID: datatypes.UUID(datatypes.BinUUIDFromString(sessionID))})
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{Data: session})
}

type ConnectToSpaceReq struct {
	SpaceID string `form:"space_id" json:"space_id" binding:"required,uuid" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ConnectToSpace godoc
//
//	@Summary		Connect session to space
//	@Description	Connect a session to a space by id
//	@Tags			session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string						true	"Session ID"	format(uuid)
//	@Param			payload		body		handler.ConnectToSpaceReq	true	"ConnectToSpace payload"
//	@Success		200			{object}	serializer.Response{}
//	@Router			/session/{session_id}/connect_to_space [post]
func (h *SessionHandler) ConnectToSpace(c *gin.Context) {
	req := ConnectToSpaceReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	sessionID := c.Param("session_id")
	if err := h.svc.UpdateByID(c.Request.Context(), &model.Session{
		ID:      datatypes.UUID(datatypes.BinUUIDFromString(sessionID)),
		SpaceID: datatypes.UUID(datatypes.BinUUIDFromString(req.SpaceID)),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{})
}

type SendMessageReq struct {
	Role  string       `form:"role" json:"role" binding:"required" example:"user"`
	Parts []dto.PartIn `form:"parts" json:"parts" binding:"required"`
}

func (h *SessionHandler) SendMessage(c *gin.Context) {
	req := SendMessageReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	sessionID := c.Param("session_id")
	ct := c.ContentType()
	fileMap := map[string]*multipart.FileHeader{}
	if strings.HasPrefix(ct, "multipart/form-data") {
		for _, p := range req.Parts {
			if p.FileField != "" {
				fh, err := c.FormFile(p.FileField)
				if err != nil {
					c.JSON(http.StatusBadRequest, serializer.ParamErr(fmt.Sprintf("missing file %s", p.FileField), err))
					return
				}
				fileMap[p.FileField] = fh
			}
		}
	}

	out, err := h.svc.SendMessage(c.Request.Context(), service.SendMessageInput{
		SessionID: datatypes.UUID(datatypes.BinUUIDFromString(sessionID)),
		Role:      req.Role,
		Parts:     req.Parts,
		Files:     fileMap,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusCreated, serializer.Response{Data: out})
}
