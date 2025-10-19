package converter

import (
	"fmt"

	"github.com/memodb-io/Acontext/internal/modules/service"
)

// AnthropicNormalizer normalizes Anthropic format to internal format
type AnthropicNormalizer struct{}

// Normalize converts Anthropic format messages to internal format
// Anthropic only has "user" and "assistant" roles
func (n *AnthropicNormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	return "", nil, fmt.Errorf("use NormalizeFromAnthropic instead")
}

// NormalizeFromAnthropic converts Anthropic specific parts to internal format
func (n *AnthropicNormalizer) NormalizeFromAnthropic(role string, parts []AnthropicPartIn) (string, []service.PartIn, error) {
	// Anthropic only has "user" and "assistant" roles
	// System messages should be sent via system parameter, not as messages
	validRoles := map[string]bool{
		"user": true, "assistant": true,
	}
	if !validRoles[role] {
		return "", nil, fmt.Errorf("invalid Anthropic role: %s (only 'user' and 'assistant' are supported)", role)
	}

	// Convert Anthropic parts to internal format
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

		case "image":
			if part.Source == nil {
				return "", nil, fmt.Errorf("image part requires source field")
			}
			meta := map[string]interface{}{
				"type":       part.Source.Type,
				"media_type": part.Source.MediaType,
			}
			if part.Source.Type == "base64" {
				meta["data"] = part.Source.Data
			} else if part.Source.Type == "url" {
				meta["url"] = part.Source.URL
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type:      "image",
				Meta:      meta,
				FileField: part.FileField,
			})

		case "tool_use":
			// Anthropic's "tool_use" -> internal "tool-call"
			if part.ID == "" || part.Name == "" {
				return "", nil, fmt.Errorf("tool_use part requires id and name fields")
			}
			meta := map[string]interface{}{
				"id":        part.ID,
				"tool_name": part.Name,
				"arguments": part.Input,
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type: "tool-call",
				Meta: meta,
			})

		case "tool_result":
			// Anthropic's "tool_result"
			if part.ToolUseID == "" {
				return "", nil, fmt.Errorf("tool_result part requires tool_use_id field")
			}
			meta := map[string]interface{}{
				"tool_call_id": part.ToolUseID,
				"result":       part.Content,
			}
			if part.IsError {
				meta["is_error"] = true
			}
			normalizedParts = append(normalizedParts, service.PartIn{
				Type: "tool-result",
				Meta: meta,
			})

		default:
			return "", nil, fmt.Errorf("unsupported Anthropic part type: %s", part.Type)
		}
	}

	return role, normalizedParts, nil
}
