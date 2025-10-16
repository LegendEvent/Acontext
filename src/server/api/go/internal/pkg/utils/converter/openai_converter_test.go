package converter

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIConverter_SimpleText(t *testing.T) {
	messages := []model.Message{
		{
			ID:        uuid.New(),
			SessionID: uuid.New(),
			Role:      "user",
			Parts: []model.Part{
				{
					Type: "text",
					Text: "Hello",
				},
			},
		},
	}

	converter := &OpenAIConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	openaiMsgs, ok := result.([]OpenAIMessage)
	require.True(t, ok)
	require.Len(t, openaiMsgs, 1)

	assert.Equal(t, "user", openaiMsgs[0].Role)
	assert.Equal(t, "Hello", openaiMsgs[0].Content)
}

func TestOpenAIConverter_MultiplePartsWithImage(t *testing.T) {
	messages := createTestMessages()
	publicURLs := createTestPublicURLs()

	converter := &OpenAIConverter{}
	result, err := converter.Convert(messages, publicURLs)
	require.NoError(t, err)

	openaiMsgs, ok := result.([]OpenAIMessage)
	require.True(t, ok)
	require.Len(t, openaiMsgs, 3)

	// Check third message with image
	assert.Equal(t, "user", openaiMsgs[2].Role)
	contentParts, ok := openaiMsgs[2].Content.([]OpenAIContentPart)
	require.True(t, ok)
	require.Len(t, contentParts, 2)

	assert.Equal(t, "text", contentParts[0].Type)
	assert.Equal(t, "Can you analyze this image?", contentParts[0].Text)

	assert.Equal(t, "image_url", contentParts[1].Type)
	require.NotNil(t, contentParts[1].ImageURL)
	assert.Equal(t, "https://example.com/test.png", contentParts[1].ImageURL.URL)
}

func TestOpenAIConverter_ToolCall(t *testing.T) {
	messages := []model.Message{
		{
			ID:        uuid.New(),
			SessionID: uuid.New(),
			Role:      "assistant",
			Parts: []model.Part{
				{
					Type: "tool-call",
					Meta: map[string]interface{}{
						"id":        "call_123",
						"tool_name": "get_weather",
						"arguments": map[string]interface{}{
							"location": "San Francisco",
						},
					},
				},
			},
		},
	}

	converter := &OpenAIConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	openaiMsgs, ok := result.([]OpenAIMessage)
	require.True(t, ok)
	require.Len(t, openaiMsgs, 1)

	assert.Equal(t, "assistant", openaiMsgs[0].Role)
	require.Len(t, openaiMsgs[0].ToolCalls, 1)

	toolCall := openaiMsgs[0].ToolCalls[0]
	assert.Equal(t, "call_123", toolCall.ID)
	assert.Equal(t, "function", toolCall.Type)
	assert.Equal(t, "get_weather", toolCall.Function.Name)

	// Verify arguments are JSON
	var args map[string]interface{}
	err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
	require.NoError(t, err)
	assert.Equal(t, "San Francisco", args["location"])
}

func TestOpenAIConverter_RoleConversion(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{"user role", "user", "user"},
		{"assistant role", "assistant", "assistant"},
		{"system role", "system", "system"},
		{"tool role", "tool", "tool"},
		{"function role", "function", "function"},
		{"unknown role", "unknown", "user"},
	}

	converter := &OpenAIConverter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.convertRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenAIConverter_ToolResult(t *testing.T) {
	messages := []model.Message{
		{
			ID:        uuid.New(),
			SessionID: uuid.New(),
			Role:      "user",
			Parts: []model.Part{
				{
					Type: "tool-result",
					Meta: map[string]interface{}{
						"tool_call_id": "call_123",
						"result":       "The weather is sunny",
					},
				},
			},
		},
	}

	converter := &OpenAIConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	openaiMsgs, ok := result.([]OpenAIMessage)
	require.True(t, ok)
	require.Len(t, openaiMsgs, 1)

	// Tool results are converted to text content
	// Could be either string or array depending on conversion logic
	assert.NotNil(t, openaiMsgs[0].Content)
}
