package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
	"github.com/memodb-io/Acontext/internal/modules/service"
)

type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{svc: s}
}

type GetTasksReq struct {
	Limit  int    `form:"limit,default=20" json:"limit" binding:"required,min=1,max=200" example:"20"`
	Cursor string `form:"cursor" json:"cursor" example:"cHJvdGVjdGVkIHZlcnNpb24gdG8gYmUgZXhjbHVkZWQgaW4gcGFyc2luZyB0aGUgY3Vyc29y"`
}

// GetTasks godoc
//
//	@Summary		Get tasks from session
//	@Description	Get tasks from session with cursor-based pagination
//	@Tags			task
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path	string	true	"Session ID"	format(uuid)
//	@Param			limit		query	integer	false	"Limit of tasks to return, default 20. Max 200."
//	@Param			cursor		query	string	false	"Cursor for pagination. Use the cursor from the previous response to get the next page."
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{data=service.GetTasksOutput}
//	@Router			/session/{session_id}/task [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	req := GetTasksReq{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	out, err := h.svc.GetTasks(c.Request.Context(), service.GetTasksInput{
		SessionID: sessionID,
		Limit:     req.Limit,
		Cursor:    req.Cursor,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.DBErr("", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{Data: out})
}
