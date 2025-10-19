package converter

import (
	"testing"

	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicNormalizer_Normalize(t *testing.T) {
	normalizer := &AnthropicNormalizer{}

	// Test that the old Normalize method returns an error
	_, _, err := normalizer.Normalize("user", []service.PartIn{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use NormalizeFromAnthropic instead")
}

func TestAnthropicNormalizer_NormalizeFromAnthropic(t *testing.T) {
	normalizer := &AnthropicNormalizer{}

	tests := []struct {
		name     string
		role     string
		parts    []AnthropicPartIn
		wantRole string
		wantErr  bool
		validate func(t *testing.T, parts []service.PartIn)
	}{
		{
			name: "valid text message",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Hello"},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "text", parts[0].Type)
				assert.Equal(t, "Hello", parts[0].Text)
			},
		},
		{
			name: "valid assistant message",
			role: "assistant",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Hi there"},
			},
			wantRole: "assistant",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "text", parts[0].Type)
			},
		},
		{
			name: "multiple text parts",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "First part"},
				{Type: "text", Text: "Second part"},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 2)
				assert.Equal(t, "First part", parts[0].Text)
				assert.Equal(t, "Second part", parts[1].Text)
			},
		},
		{
			name: "image message with base64",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type: "image",
					Source: &AnthropicImageSource{
						Type:      "base64",
						MediaType: "image/jpeg",
						Data:      "base64data...",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "image", parts[0].Type)
				assert.NotNil(t, parts[0].Meta)
				assert.Equal(t, "base64", parts[0].Meta["type"])
				assert.Equal(t, "image/jpeg", parts[0].Meta["media_type"])
				assert.Equal(t, "base64data...", parts[0].Meta["data"])
			},
		},
		{
			name: "image message with url",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type: "image",
					Source: &AnthropicImageSource{
						Type:      "url",
						MediaType: "image/png",
						URL:       "https://example.com/image.png",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "image", parts[0].Type)
				assert.Equal(t, "url", parts[0].Meta["type"])
				assert.Equal(t, "https://example.com/image.png", parts[0].Meta["url"])
			},
		},
		{
			name: "tool_use message",
			role: "assistant",
			parts: []AnthropicPartIn{
				{
					Type:  "tool_use",
					ID:    "toolu_abc123",
					Name:  "get_weather",
					Input: map[string]interface{}{"city": "San Francisco"},
				},
			},
			wantRole: "assistant",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-call", parts[0].Type) // tool_use converts to tool-call
				assert.Equal(t, "toolu_abc123", parts[0].Meta["id"])
				assert.Equal(t, "get_weather", parts[0].Meta["tool_name"])

				args, ok := parts[0].Meta["arguments"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "San Francisco", args["city"])
			},
		},
		{
			name: "tool_use with nil input",
			role: "assistant",
			parts: []AnthropicPartIn{
				{
					Type:  "tool_use",
					ID:    "toolu_123",
					Name:  "simple_tool",
					Input: nil,
				},
			},
			wantRole: "assistant",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-call", parts[0].Type)
				assert.Nil(t, parts[0].Meta["arguments"])
			},
		},
		{
			name: "tool_result message",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type:      "tool_result",
					ToolUseID: "toolu_abc123",
					Content:   "Sunny, 72°F",
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-result", parts[0].Type)
				assert.Equal(t, "toolu_abc123", parts[0].Meta["tool_call_id"])
				assert.Equal(t, "Sunny, 72°F", parts[0].Meta["result"])
			},
		},
		{
			name: "tool_result with error flag",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type:      "tool_result",
					ToolUseID: "toolu_456",
					Content:   "Error occurred",
					IsError:   true,
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-result", parts[0].Type)
				assert.Equal(t, true, parts[0].Meta["is_error"])
			},
		},
		{
			name: "tool_result without error flag",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type:      "tool_result",
					ToolUseID: "toolu_789",
					Content:   "Success",
					IsError:   false,
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.NotContains(t, parts[0].Meta, "is_error")
			},
		},
		{
			name: "text with file field",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Hello", FileField: "file1"},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "file1", parts[0].FileField)
			},
		},
		{
			name: "mixed parts",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Look at this:"},
				{
					Type: "image",
					Source: &AnthropicImageSource{
						Type:      "url",
						MediaType: "image/jpeg",
						URL:       "https://example.com/img.jpg",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 2)
				assert.Equal(t, "text", parts[0].Type)
				assert.Equal(t, "image", parts[1].Type)
			},
		},

		// Error cases
		{
			name: "system role is invalid for Anthropic",
			role: "system",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "You are a helpful assistant"},
			},
			wantErr: true,
		},
		{
			name: "tool role is invalid for Anthropic",
			role: "tool",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Result"},
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			role: "invalid_role",
			parts: []AnthropicPartIn{
				{Type: "text", Text: "Hello"},
			},
			wantErr: true,
		},
		{
			name: "empty text should error",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "text", Text: ""},
			},
			wantErr: true,
		},
		{
			name: "image without source",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "image"},
			},
			wantErr: true,
		},
		{
			name: "tool_use without ID should error",
			role: "assistant",
			parts: []AnthropicPartIn{
				{
					Type:  "tool_use",
					Name:  "get_weather",
					Input: map[string]interface{}{"city": "SF"},
				},
			},
			wantErr: true,
		},
		{
			name: "tool_use without Name should error",
			role: "assistant",
			parts: []AnthropicPartIn{
				{
					Type:  "tool_use",
					ID:    "toolu_123",
					Input: map[string]interface{}{"city": "SF"},
				},
			},
			wantErr: true,
		},
		{
			name: "tool_result without tool_use_id",
			role: "user",
			parts: []AnthropicPartIn{
				{
					Type:    "tool_result",
					Content: "Result",
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported part type",
			role: "user",
			parts: []AnthropicPartIn{
				{Type: "unknown_type"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedRole, normalizedParts, err := normalizer.NormalizeFromAnthropic(tt.role, tt.parts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRole, normalizedRole)

			if tt.validate != nil {
				tt.validate(t, normalizedParts)
			}
		})
	}
}

func TestAnthropicNormalizer_ComplexScenarios(t *testing.T) {
	normalizer := &AnthropicNormalizer{}

	t.Run("conversation with tool use and result", func(t *testing.T) {
		// First message: assistant with tool_use
		role1, parts1, err := normalizer.NormalizeFromAnthropic("assistant", []AnthropicPartIn{
			{Type: "text", Text: "Let me check the weather for you."},
			{
				Type:  "tool_use",
				ID:    "toolu_01A",
				Name:  "get_weather",
				Input: map[string]interface{}{"location": "New York"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "assistant", role1)
		assert.Len(t, parts1, 2)
		assert.Equal(t, "text", parts1[0].Type)
		assert.Equal(t, "tool-call", parts1[1].Type)

		// Second message: user with tool_result
		role2, parts2, err := normalizer.NormalizeFromAnthropic("user", []AnthropicPartIn{
			{
				Type:      "tool_result",
				ToolUseID: "toolu_01A",
				Content:   "Temperature: 68°F, Condition: Partly cloudy",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "user", role2)
		assert.Len(t, parts2, 1)
		assert.Equal(t, "tool-result", parts2[0].Type)
	})

	t.Run("multiple images in one message", func(t *testing.T) {
		role, parts, err := normalizer.NormalizeFromAnthropic("user", []AnthropicPartIn{
			{Type: "text", Text: "Compare these two images:"},
			{
				Type: "image",
				Source: &AnthropicImageSource{
					Type:      "url",
					MediaType: "image/jpeg",
					URL:       "https://example.com/img1.jpg",
				},
			},
			{
				Type: "image",
				Source: &AnthropicImageSource{
					Type:      "url",
					MediaType: "image/jpeg",
					URL:       "https://example.com/img2.jpg",
				},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "user", role)
		assert.Len(t, parts, 3)
		assert.Equal(t, "text", parts[0].Type)
		assert.Equal(t, "image", parts[1].Type)
		assert.Equal(t, "image", parts[2].Type)
	})
}
