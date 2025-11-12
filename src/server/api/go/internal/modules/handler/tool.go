package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memodb-io/Acontext/internal/infra/httpclient"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/serializer"
)

type ToolHandler struct {
	coreClient *httpclient.CoreClient
}

func NewToolHandler(coreClient *httpclient.CoreClient) *ToolHandler {
	return &ToolHandler{
		coreClient: coreClient,
	}
}

type ToolRenameItem struct {
	OldName string `json:"old_name" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

type RenameToolNameReq struct {
	Rename []ToolRenameItem `json:"rename" binding:"required,min=1"`
}

// RenameToolName godoc
//
//	@Summary		Rename tool names
//	@Description	Rename one or more tool names within a project
//	@Tags			tool
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	handler.RenameToolNameReq	true	"Tool rename request"
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{data=httpclient.FlagResponse}
//	@Router			/tool/name [put]
//	@x-code-samples	[{"lang":"python","source":"from acontext import AcontextClient\n\nclient = AcontextClient(api_key='sk_project_token')\n\n# Rename tool names\nresult = client.tools.rename([\n    {\"old_name\": \"old_tool_name\", \"new_name\": \"new_tool_name\"}\n])\nprint(result.status)\n","label":"Python"},{"lang":"javascript","source":"import { AcontextClient } from '@acontext/acontext';\n\nconst client = new AcontextClient({ apiKey: 'sk_project_token' });\n\n// Rename tool names\nconst result = await client.tools.rename([\n  { oldName: 'old_tool_name', newName: 'new_tool_name' }\n]);\nconsole.log(result.status);\n","label":"JavaScript"}]
func (h *ToolHandler) RenameToolName(c *gin.Context) {
	// Get project from context
	project, ok := c.MustGet("project").(*model.Project)
	if !ok {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", errors.New("project not found")))
		return
	}

	req := RenameToolNameReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", err))
		return
	}

	// Convert handler types to httpclient types
	renameItems := make([]httpclient.ToolRenameItem, len(req.Rename))
	for i, item := range req.Rename {
		renameItems[i] = httpclient.ToolRenameItem{
			OldName: item.OldName,
			NewName: item.NewName,
		}
	}

	// Call Core service to rename tools
	result, err := h.coreClient.ToolRename(c.Request.Context(), project.ID, renameItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.Err(http.StatusInternalServerError, "failed to rename tools", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{Data: result})
}

// GetToolName godoc
//
//	@Summary		Get tool names
//	@Description	Get all tool names within a project
//	@Tags			tool
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{data=[]httpclient.ToolReferenceData}
//	@Router			/tool/name [get]
//	@x-code-samples	[{"lang":"python","source":"from acontext import AcontextClient\n\nclient = AcontextClient(api_key='sk_project_token')\n\n# Get all tool names\ntools = client.tools.list()\nfor tool in tools:\n    print(f\"{tool.name}: {tool.sop_count} SOPs\")\n","label":"Python"},{"lang":"javascript","source":"import { AcontextClient } from '@acontext/acontext';\n\nconst client = new AcontextClient({ apiKey: 'sk_project_token' });\n\n// Get all tool names\nconst tools = await client.tools.list();\nfor (const tool of tools) {\n  console.log(`${tool.name}: ${tool.sop_count} SOPs`);\n}\n","label":"JavaScript"}]
func (h *ToolHandler) GetToolName(c *gin.Context) {
	// Get project from context
	project, ok := c.MustGet("project").(*model.Project)
	if !ok {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("", errors.New("project not found")))
		return
	}

	// Call Core service to get tool names
	result, err := h.coreClient.GetToolNames(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.Err(http.StatusInternalServerError, "failed to get tool names", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Response{Data: result})
}
