package converter

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/service"
)

// Anthropic message structures compatible with Claude Messages API
// Reference: https://docs.claude.com/en/api/messages

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []AnthropicContentBlock
}

type AnthropicContentBlock struct {
	Type string `json:"type"` // "text", "image", "tool_use", "tool_result"

	// For text blocks
	Text string `json:"text,omitempty"`

	// For image blocks
	Source *AnthropicImageSource `json:"source,omitempty"`

	// For tool_use blocks
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`

	// For tool_result blocks
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"` // string or nested content blocks
	IsError   bool        `json:"is_error,omitempty"`
}

type AnthropicImageSource struct {
	Type      string `json:"type"` // "base64" or "url"
	MediaType string `json:"media_type"`
	Data      string `json:"data,omitempty"` // base64 data
	URL       string `json:"url,omitempty"`  // image URL
}

// AnthropicConverter converts messages to Anthropic Claude-compatible format
type AnthropicConverter struct{}

func (c *AnthropicConverter) Convert(messages []model.Message, publicURLs map[string]service.PublicURL) (interface{}, error) {
	result := make([]AnthropicMessage, 0, len(messages))

	for _, msg := range messages {
		// Skip system messages - they should be handled separately via system parameter
		if msg.Role == "system" {
			continue
		}

		anthropicMsg := AnthropicMessage{
			Role: c.convertRole(msg.Role),
		}

		// Convert parts to content
		if len(msg.Parts) > 0 {
			content := c.convertParts(msg.Parts, publicURLs)
			anthropicMsg.Content = content
		}

		result = append(result, anthropicMsg)
	}

	return result, nil
}

func (c *AnthropicConverter) convertRole(role string) string {
	// Anthropic roles: "user", "assistant"
	// Note: "system" messages should be passed via the top-level system parameter
	switch role {
	case "assistant":
		return "assistant"
	case "user", "tool", "function":
		return "user"
	default:
		return "user"
	}
}

func (c *AnthropicConverter) convertParts(parts []model.Part, publicURLs map[string]service.PublicURL) interface{} {
	// If only one text part, return as string
	if len(parts) == 1 && parts[0].Type == "text" {
		return parts[0].Text
	}

	// Multiple parts or non-text parts, convert to content blocks array
	contentBlocks := make([]AnthropicContentBlock, 0, len(parts))

	for _, part := range parts {
		switch part.Type {
		case "text":
			if part.Text != "" {
				contentBlocks = append(contentBlocks, AnthropicContentBlock{
					Type: "text",
					Text: part.Text,
				})
			}

		case "image":
			imageBlock := c.convertImagePart(part, publicURLs)
			if imageBlock != nil {
				contentBlocks = append(contentBlocks, *imageBlock)
			}

		case "tool-call":
			toolUseBlock := c.convertToolCallPart(part)
			if toolUseBlock != nil {
				contentBlocks = append(contentBlocks, *toolUseBlock)
			}

		case "tool-result":
			toolResultBlock := c.convertToolResultPart(part)
			if toolResultBlock != nil {
				contentBlocks = append(contentBlocks, *toolResultBlock)
			}

		case "audio", "video", "file":
			// For media files not supported by Anthropic, add as text reference
			assetURL := c.getAssetURL(part.Asset, publicURLs)
			text := part.Text
			if assetURL != "" {
				text = fmt.Sprintf("%s\n[%s: %s]", text, part.Type, assetURL)
			}
			if text != "" {
				contentBlocks = append(contentBlocks, AnthropicContentBlock{
					Type: "text",
					Text: text,
				})
			}

		default:
			// For other types, include as text if available
			if part.Text != "" {
				contentBlocks = append(contentBlocks, AnthropicContentBlock{
					Type: "text",
					Text: part.Text,
				})
			}
		}
	}

	// If only one text block, return as string
	if len(contentBlocks) == 1 && contentBlocks[0].Type == "text" {
		return contentBlocks[0].Text
	}

	return contentBlocks
}

func (c *AnthropicConverter) convertImagePart(part model.Part, publicURLs map[string]service.PublicURL) *AnthropicContentBlock {
	if part.Asset == nil {
		return nil
	}

	imageURL := c.getAssetURL(part.Asset, publicURLs)
	if imageURL == "" {
		return nil
	}

	// Determine media type
	mediaType := part.Asset.MIME
	if mediaType == "" {
		mediaType = "image/jpeg" // default
	}

	// Try to fetch and convert to base64 if it's a public URL
	// Anthropic prefers base64 encoded images
	base64Data := c.fetchAndEncodeImage(imageURL, mediaType)

	block := &AnthropicContentBlock{
		Type: "image",
	}

	if base64Data != "" {
		block.Source = &AnthropicImageSource{
			Type:      "base64",
			MediaType: mediaType,
			Data:      base64Data,
		}
	} else {
		// Fallback to URL if base64 conversion fails
		block.Source = &AnthropicImageSource{
			Type:      "url",
			MediaType: mediaType,
			URL:       imageURL,
		}
	}

	return block
}

func (c *AnthropicConverter) convertToolCallPart(part model.Part) *AnthropicContentBlock {
	if part.Meta == nil {
		return nil
	}

	block := &AnthropicContentBlock{
		Type: "tool_use",
	}

	if id, ok := part.Meta["id"].(string); ok {
		block.ID = id
	}
	if name, ok := part.Meta["tool_name"].(string); ok {
		block.Name = name
	}
	if args, ok := part.Meta["arguments"].(map[string]interface{}); ok {
		block.Input = args
	} else if argsRaw := part.Meta["arguments"]; argsRaw != nil {
		// Try to convert to map
		if argsBytes, err := sonic.Marshal(argsRaw); err == nil {
			var argsMap map[string]interface{}
			if err := sonic.Unmarshal(argsBytes, &argsMap); err == nil {
				block.Input = argsMap
			}
		}
	}

	return block
}

func (c *AnthropicConverter) convertToolResultPart(part model.Part) *AnthropicContentBlock {
	if part.Meta == nil {
		return nil
	}

	block := &AnthropicContentBlock{
		Type: "tool_result",
	}

	if toolCallID, ok := part.Meta["tool_call_id"].(string); ok {
		block.ToolUseID = toolCallID
	} else if id, ok := part.Meta["id"].(string); ok {
		block.ToolUseID = id
	}

	// Get content from meta or text
	var content interface{}
	if result, ok := part.Meta["result"]; ok {
		content = result
	} else if part.Text != "" {
		content = part.Text
	} else {
		// Use the entire meta as content
		content = part.Meta
	}

	// Convert to string if it's not already
	if str, ok := content.(string); ok {
		block.Content = str
	} else {
		if jsonBytes, err := sonic.Marshal(content); err == nil {
			block.Content = string(jsonBytes)
		}
	}

	// Check for error flag
	if isError, ok := part.Meta["is_error"].(bool); ok {
		block.IsError = isError
	}

	return block
}

func (c *AnthropicConverter) getAssetURL(asset *model.Asset, publicURLs map[string]service.PublicURL) string {
	if asset == nil {
		return ""
	}
	if pubURL, ok := publicURLs[asset.SHA256]; ok {
		return pubURL.URL
	}
	return ""
}

func (c *AnthropicConverter) fetchAndEncodeImage(imageURL, mediaType string) string {
	// Only try to fetch if it's an HTTP(S) URL
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		return ""
	}

	// Make HTTP request to fetch image
	resp, err := http.Get(imageURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	// Read image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(data)
}
