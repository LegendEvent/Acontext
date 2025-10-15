package converter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		wantErr   bool
		wantValue MessageFormat
	}{
		{
			name:      "empty format",
			format:    "",
			wantErr:   false,
			wantValue: FormatNone,
		},
		{
			name:      "openai format",
			format:    "openai",
			wantErr:   false,
			wantValue: FormatOpenAI,
		},
		{
			name:      "langchain format",
			format:    "langchain",
			wantErr:   false,
			wantValue: FormatLangChain,
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func createTestMessages() []model.Message {
	sessionID := uuid.New()

	return []model.Message{
		{
			ID:        uuid.New(),
			SessionID: sessionID,
			Role:      "user",
			Parts: []model.Part{
				{
					Type: "text",
					Text: "Hello, how are you?",
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			SessionID: sessionID,
			Role:      "assistant",
			Parts: []model.Part{
				{
					Type: "text",
					Text: "I'm doing well, thank you! How can I help you today?",
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			SessionID: sessionID,
			Role:      "user",
			Parts: []model.Part{
				{
					Type: "text",
					Text: "Can you analyze this image?",
				},
				{
					Type: "image",
					Asset: &model.Asset{
						SHA256: "abc123",
						MIME:   "image/png",
					},
					Filename: "test.png",
				},
			},
			CreatedAt: time.Now(),
		},
	}
}

func createTestPublicURLs() map[string]service.PublicURL {
	return map[string]service.PublicURL{
		"abc123": {
			URL:      "https://example.com/test.png",
			ExpireAt: time.Now().Add(24 * time.Hour),
		},
	}
}

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

func TestConvertMessages(t *testing.T) {
	messages := createTestMessages()
	publicURLs := createTestPublicURLs()

	tests := []struct {
		name   string
		format MessageFormat
	}{
		{
			name:   "no conversion",
			format: FormatNone,
		},
		{
			name:   "openai conversion",
			format: FormatOpenAI,
		},
		{
			name:   "langchain conversion",
			format: FormatLangChain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertMessages(ConvertMessagesInput{
				Messages:   messages,
				Format:     tt.format,
				PublicURLs: publicURLs,
			})
			require.NoError(t, err)
			assert.NotNil(t, result)

			switch tt.format {
			case FormatNone:
				_, ok := result.([]model.Message)
				assert.True(t, ok)
			case FormatOpenAI:
				_, ok := result.([]OpenAIMessage)
				assert.True(t, ok)
			case FormatLangChain:
				// LangChain returns []llms.ChatMessage
				_, ok := result.([]llms.ChatMessage)
				assert.True(t, ok)
			}
		})
	}
}

func TestGetConvertedMessagesOutput(t *testing.T) {
	messages := createTestMessages()
	publicURLs := createTestPublicURLs()

	result, err := GetConvertedMessagesOutput(
		messages,
		FormatOpenAI,
		publicURLs,
		"next_cursor_value",
		true,
	)
	require.NoError(t, err)

	assert.Contains(t, result, "items")
	assert.Contains(t, result, "has_more")
	assert.Contains(t, result, "next_cursor")
	assert.Equal(t, true, result["has_more"])
	assert.Equal(t, "next_cursor_value", result["next_cursor"])

	// Public URLs should not be included for non-None formats
	assert.NotContains(t, result, "public_urls")
}

func TestGetConvertedMessagesOutput_NoneFormat(t *testing.T) {
	messages := createTestMessages()
	publicURLs := createTestPublicURLs()

	result, err := GetConvertedMessagesOutput(
		messages,
		FormatNone,
		publicURLs,
		"",
		false,
	)
	require.NoError(t, err)

	assert.Contains(t, result, "items")
	assert.Contains(t, result, "has_more")
	assert.NotContains(t, result, "next_cursor")
	assert.Equal(t, false, result["has_more"])

	// Public URLs should be included for None format
	assert.Contains(t, result, "public_urls")
}

func TestConvertMessages_UnsupportedFormat(t *testing.T) {
	messages := createTestMessages()

	_, err := ConvertMessages(ConvertMessagesInput{
		Messages:   messages,
		Format:     MessageFormat("unsupported"),
		PublicURLs: nil,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}
