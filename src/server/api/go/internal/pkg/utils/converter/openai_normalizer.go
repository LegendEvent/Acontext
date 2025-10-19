package converter

import (
	"encoding/json"
	"fmt"

	"github.com/memodb-io/Acontext/internal/modules/service"
)

// OpenAINormalizer normalizes OpenAI format to internal format
type OpenAINormalizer struct{}

// Normalize converts OpenAI format messages to internal format
// OpenAI supports: user, assistant, system, tool, function
// Internal format only supports: user, assistant, system
func (n *OpenAINormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	return "", nil, fmt.Errorf("use NormalizeFromOpenAI instead")
}

// NormalizeFromOpenAI converts OpenAI specific parts to internal format
func (n *OpenAINormalizer) NormalizeFromOpenAI(role string, parts []OpenAIPartIn) (string, []service.PartIn, error) {
	normalizedRole := role

	// Convert deprecated/external roles to internal format
	if role == "tool" || role == "function" {
		normalizedRole = "user"
	}

	// Validate normalized role
	validRoles := map[string]bool{
		"user": true, "assistant": true, "system": true,
	}
	if !validRoles[normalizedRole] {
		return "", nil, fmt.Errorf("invalid OpenAI role: %s", role)
	}

	// Convert OpenAI parts to internal format
	normalizedParts := make([]service.PartIn, 0, len(parts))

	for _, part := range parts {
		switch part.Type {
		case "text":
			if part.Text == "" {
				return "", nil, fmt.Errorf("text part requires non-empty text field")
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type:      "text",
				Text:      part.Text,
				FileField: part.FileField,
			})

		case "image_url":
			if part.ImageURL == nil {
				return "", nil, fmt.Errorf("image_url part requires image_url field")
			}
			meta := map[string]interface{}{
				"url": part.ImageURL.URL,
			}
			if part.ImageURL.Detail != "" {
				meta["detail"] = part.ImageURL.Detail
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type:      "image",
				Meta:      meta,
				FileField: part.FileField,
			})

		case "input_audio":
			if part.InputAudio == nil {
				return "", nil, fmt.Errorf("input_audio part requires input_audio field")
			}
			meta := map[string]interface{}{
				"data":   part.InputAudio.Data,
				"format": part.InputAudio.Format,
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type: "audio",
				Meta: meta,
			})

		case "tool_call":
			// OpenAI tool_call (from assistant)
			if part.ID == "" || part.Function == nil {
				return "", nil, fmt.Errorf("tool_call part requires id and function fields")
			}

			// Parse arguments from JSON string
			var arguments map[string]interface{}
			if part.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(part.Function.Arguments), &arguments); err != nil {
					return "", nil, fmt.Errorf("invalid tool_call arguments JSON: %w", err)
				}
			}

			meta := map[string]interface{}{
				"id":        part.ID,
				"tool_name": part.Function.Name,
				"arguments": arguments,
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type: "tool-call",
				Meta: meta,
			})

		case "tool_result":
			// OpenAI tool result (from user)
			if part.ToolCallID == "" {
				return "", nil, fmt.Errorf("tool_result part requires tool_call_id field")
			}
			meta := map[string]interface{}{
				"tool_call_id": part.ToolCallID,
				"result":       part.Output,
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type: "tool-result",
				Meta: meta,
			})

		default:
			return "", nil, fmt.Errorf("unsupported OpenAI part type: %s", part.Type)
		}
	}

	return normalizedRole, normalizedParts, nil
}
