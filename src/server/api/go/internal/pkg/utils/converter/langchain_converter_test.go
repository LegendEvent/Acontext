package converter

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestLangChainConverter_SimpleText(t *testing.T) {
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

	converter := &LangChainConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	langchainMsgs, ok := result.([]llms.ChatMessage)
	require.True(t, ok)
	require.Len(t, langchainMsgs, 1)

	// Verify message content
	assert.Equal(t, "Hello", langchainMsgs[0].GetContent())
	assert.Equal(t, llms.ChatMessageTypeHuman, langchainMsgs[0].GetType())
}

func TestLangChainConverter_RoleConversion(t *testing.T) {
	tests := []struct {
		role string
	}{
		{role: "user"},
		{role: "assistant"},
		{role: "system"},
		{role: "tool"},
		{role: "function"},
	}

	converter := &LangChainConverter{}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			messages := []model.Message{
				{
					ID:        uuid.New(),
					SessionID: uuid.New(),
					Role:      tt.role,
					Parts: []model.Part{
						{
							Type: "text",
							Text: "test",
						},
					},
				},
			}

			result, err := converter.Convert(messages, nil)
			require.NoError(t, err)

			langchainMsgs, ok := result.([]llms.ChatMessage)
			require.True(t, ok)
			require.Len(t, langchainMsgs, 1)

			// Verify message content
			assert.Equal(t, "test", langchainMsgs[0].GetContent())
		})
	}
}

func TestLangChainConverter_MultipleParts(t *testing.T) {
	messages := createTestMessages()
	publicURLs := createTestPublicURLs()

	converter := &LangChainConverter{}
	result, err := converter.Convert(messages, publicURLs)
	require.NoError(t, err)

	langchainMsgs, ok := result.([]llms.ChatMessage)
	require.True(t, ok)
	require.Len(t, langchainMsgs, 3)

	// Check third message with image - content should be JSON array
	content := langchainMsgs[2].GetContent()

	var contentParts []map[string]interface{}
	err = json.Unmarshal([]byte(content), &contentParts)
	require.NoError(t, err)
	require.Len(t, contentParts, 2)

	assert.Equal(t, "text", contentParts[0]["type"])
	assert.Equal(t, "Can you analyze this image?", contentParts[0]["text"])

	assert.Equal(t, "image", contentParts[1]["type"])
	assert.Equal(t, "https://example.com/test.png", contentParts[1]["url"])
	assert.Equal(t, "test.png", contentParts[1]["filename"])
}

func TestLangChainConverter_ToolCalls(t *testing.T) {
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

	converter := &LangChainConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	langchainMsgs, ok := result.([]llms.ChatMessage)
	require.True(t, ok)
	require.Len(t, langchainMsgs, 1)

	// Check if it's an AI message with tool calls
	aiMsg, ok := langchainMsgs[0].(llms.AIChatMessage)
	require.True(t, ok)
	require.Len(t, aiMsg.ToolCalls, 1)

	toolCall := aiMsg.ToolCalls[0]
	assert.Equal(t, "call_123", toolCall.ID)
	assert.NotNil(t, toolCall.FunctionCall)
	assert.Equal(t, "get_weather", toolCall.FunctionCall.Name)
}

func TestLangChainConverter_ToolResult(t *testing.T) {
	messages := []model.Message{
		{
			ID:        uuid.New(),
			SessionID: uuid.New(),
			Role:      "tool",
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

	converter := &LangChainConverter{}
	result, err := converter.Convert(messages, nil)
	require.NoError(t, err)

	langchainMsgs, ok := result.([]llms.ChatMessage)
	require.True(t, ok)
	require.Len(t, langchainMsgs, 1)

	// Check if it's a tool message
	toolMsg, ok := langchainMsgs[0].(llms.ToolChatMessage)
	require.True(t, ok)
	assert.Equal(t, "call_123", toolMsg.ID)
}
