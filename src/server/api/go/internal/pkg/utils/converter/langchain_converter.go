package converter

import (
	"github.com/bytedance/sonic"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/tmc/langchaingo/llms"
)

// LangChainConverter converts messages to LangChain format using official llms package types
// Reference: github.com/tmc/langchaingo/llms
type LangChainConverter struct{}

func (c *LangChainConverter) Convert(messages []model.Message, publicURLs map[string]service.PublicURL) (interface{}, error) {
	result := make([]llms.ChatMessage, 0, len(messages))

	for _, msg := range messages {
		var langchainMsg llms.ChatMessage

		// Combine all text parts into content
		content := c.extractContent(msg.Parts, publicURLs)

		switch msg.Role {
		case "user":
			langchainMsg = llms.HumanChatMessage{
				Content: content,
			}
		case "assistant":
			// Check if there are tool calls
			toolCalls := c.extractToolCalls(msg.Parts)
			langchainMsg = llms.AIChatMessage{
				Content:   content,
				ToolCalls: toolCalls,
			}
		case "function":
			langchainMsg = llms.AIChatMessage{
				Content: content,
			}
		case "system":
			langchainMsg = llms.SystemChatMessage{
				Content: content,
			}
		case "tool":
			// Extract tool call ID from parts
			toolCallID := c.extractToolCallID(msg.Parts)
			langchainMsg = llms.ToolChatMessage{
				ID:      toolCallID,
				Content: content,
			}
		default:
			langchainMsg = llms.GenericChatMessage{
				Content: content,
				Role:    msg.Role,
			}
		}

		result = append(result, langchainMsg)
	}

	return result, nil
}

func (c *LangChainConverter) extractContent(parts []model.Part, publicURLs map[string]service.PublicURL) string {
	if len(parts) == 0 {
		return ""
	}

	// If single text part, return directly
	if len(parts) == 1 && parts[0].Type == "text" {
		return parts[0].Text
	}

	// Multiple parts: combine into structured content
	contentParts := make([]map[string]interface{}, 0, len(parts))

	for _, part := range parts {
		partMap := map[string]interface{}{
			"type": part.Type,
		}

		switch part.Type {
		case "text":
			partMap["text"] = part.Text

		case "image", "audio", "video", "file":
			if part.Asset != nil {
				if pubURL, ok := publicURLs[part.Asset.SHA256]; ok {
					partMap["url"] = pubURL.URL
				}
				partMap["filename"] = part.Filename
				partMap["mime"] = part.Asset.MIME
			}
			if part.Text != "" {
				partMap["text"] = part.Text
			}

		case "tool-call", "tool-result", "data":
			partMap["meta"] = part.Meta
			if part.Text != "" {
				partMap["text"] = part.Text
			}
		}

		contentParts = append(contentParts, partMap)
	}

	// Serialize to JSON string for LangChain
	if jsonBytes, err := sonic.Marshal(contentParts); err == nil {
		return string(jsonBytes)
	}

	return ""
}

func (c *LangChainConverter) extractToolCalls(parts []model.Part) []llms.ToolCall {
	var toolCalls []llms.ToolCall

	for _, part := range parts {
		if part.Type == "tool-call" && part.Meta != nil {
			toolCall := llms.ToolCall{}

			if id, ok := part.Meta["id"].(string); ok {
				toolCall.ID = id
			}
			if toolName, ok := part.Meta["tool_name"].(string); ok {
				toolCall.FunctionCall = &llms.FunctionCall{
					Name: toolName,
				}
			}
			if args, ok := part.Meta["arguments"]; ok {
				if argsBytes, err := sonic.Marshal(args); err == nil {
					if toolCall.FunctionCall != nil {
						toolCall.FunctionCall.Arguments = string(argsBytes)
					}
				}
			}

			toolCalls = append(toolCalls, toolCall)
		}
	}

	return toolCalls
}

func (c *LangChainConverter) extractToolCallID(parts []model.Part) string {
	for _, part := range parts {
		if part.Type == "tool-result" && part.Meta != nil {
			if toolCallID, ok := part.Meta["tool_call_id"].(string); ok {
				return toolCallID
			}
		}
	}
	return ""
}
