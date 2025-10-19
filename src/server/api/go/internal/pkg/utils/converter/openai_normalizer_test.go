package converter

import (
	"testing"

	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAINormalizer_Normalize(t *testing.T) {
	normalizer := &OpenAINormalizer{}

	// Test that the old Normalize method returns an error
	_, _, err := normalizer.Normalize("user", []service.PartIn{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use NormalizeFromOpenAI instead")
}

func TestOpenAINormalizer_NormalizeFromOpenAI(t *testing.T) {
	normalizer := &OpenAINormalizer{}

	tests := []struct {
		name     string
		role     string
		parts    []OpenAIPartIn
		wantRole string
		wantErr  bool
		validate func(t *testing.T, parts []service.PartIn)
	}{
		{
			name: "valid text message",
			role: "user",
			parts: []OpenAIPartIn{
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
			name: "multiple text parts",
			role: "user",
			parts: []OpenAIPartIn{
				{Type: "text", Text: "Part 1"},
				{Type: "text", Text: "Part 2"},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 2)
				assert.Equal(t, "Part 1", parts[0].Text)
				assert.Equal(t, "Part 2", parts[1].Text)
			},
		},
		{
			name: "image_url message",
			role: "user",
			parts: []OpenAIPartIn{
				{
					Type: "image_url",
					ImageURL: &OpenAIImageURL{
						URL:    "https://example.com/image.jpg",
						Detail: "high",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "image", parts[0].Type)
				assert.NotNil(t, parts[0].Meta)
				assert.Equal(t, "https://example.com/image.jpg", parts[0].Meta["url"])
				assert.Equal(t, "high", parts[0].Meta["detail"])
			},
		},
		{
			name: "image_url without detail",
			role: "user",
			parts: []OpenAIPartIn{
				{
					Type: "image_url",
					ImageURL: &OpenAIImageURL{
						URL: "https://example.com/image.jpg",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "image", parts[0].Type)
				assert.Equal(t, "https://example.com/image.jpg", parts[0].Meta["url"])
				assert.NotContains(t, parts[0].Meta, "detail")
			},
		},
		{
			name: "input_audio message",
			role: "user",
			parts: []OpenAIPartIn{
				{
					Type: "input_audio",
					InputAudio: &OpenAIInputAudio{
						Data:   "base64audiodata",
						Format: "wav",
					},
				},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "audio", parts[0].Type)
				assert.Equal(t, "base64audiodata", parts[0].Meta["data"])
				assert.Equal(t, "wav", parts[0].Meta["format"])
			},
		},
		{
			name: "tool_call message",
			role: "assistant",
			parts: []OpenAIPartIn{
				{
					Type: "tool_call",
					ID:   "call_abc123",
					Function: &OpenAIFunctionCall{
						Name:      "get_weather",
						Arguments: `{"city":"San Francisco"}`,
					},
				},
			},
			wantRole: "assistant",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-call", parts[0].Type)
				assert.Equal(t, "call_abc123", parts[0].Meta["id"])
				assert.Equal(t, "get_weather", parts[0].Meta["tool_name"])

				args, ok := parts[0].Meta["arguments"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "San Francisco", args["city"])
			},
		},
		{
			name: "tool_call with empty arguments",
			role: "assistant",
			parts: []OpenAIPartIn{
				{
					Type: "tool_call",
					ID:   "call_123",
					Function: &OpenAIFunctionCall{
						Name:      "no_args_tool",
						Arguments: "",
					},
				},
			},
			wantRole: "assistant",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-call", parts[0].Type)
			},
		},
		{
			name: "tool_result message",
			role: "tool",
			parts: []OpenAIPartIn{
				{
					Type:       "tool_result",
					ToolCallID: "call_abc123",
					Output:     "Sunny, 72°F",
				},
			},
			wantRole: "user", // tool role converts to user
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
				assert.Equal(t, "tool-result", parts[0].Type)
				assert.Equal(t, "call_abc123", parts[0].Meta["tool_call_id"])
				assert.Equal(t, "Sunny, 72°F", parts[0].Meta["result"])
			},
		},
		{
			name: "function role converts to user",
			role: "function",
			parts: []OpenAIPartIn{
				{Type: "text", Text: "Result"},
			},
			wantRole: "user",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
			},
		},
		{
			name: "system role stays as system",
			role: "system",
			parts: []OpenAIPartIn{
				{Type: "text", Text: "You are a helpful assistant"},
			},
			wantRole: "system",
			wantErr:  false,
			validate: func(t *testing.T, parts []service.PartIn) {
				require.Len(t, parts, 1)
			},
		},
		{
			name: "text with file field",
			role: "user",
			parts: []OpenAIPartIn{
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
			parts: []OpenAIPartIn{
				{Type: "text", Text: "Look at this image:"},
				{
					Type: "image_url",
					ImageURL: &OpenAIImageURL{
						URL: "https://example.com/img.jpg",
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
			name: "invalid role",
			role: "invalid_role",
			parts: []OpenAIPartIn{
				{Type: "text", Text: "Hello"},
			},
			wantErr: true,
		},
		{
			name: "empty text should error",
			role: "user",
			parts: []OpenAIPartIn{
				{Type: "text", Text: ""},
			},
			wantErr: true,
		},
		{
			name: "image_url without ImageURL field",
			role: "user",
			parts: []OpenAIPartIn{
				{Type: "image_url"},
			},
			wantErr: true,
		},
		{
			name: "input_audio without InputAudio field",
			role: "user",
			parts: []OpenAIPartIn{
				{Type: "input_audio"},
			},
			wantErr: true,
		},
		{
			name: "tool_call without ID should error",
			role: "assistant",
			parts: []OpenAIPartIn{
				{
					Type: "tool_call",
					Function: &OpenAIFunctionCall{
						Name:      "get_weather",
						Arguments: `{"city":"SF"}`,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "tool_call without Function should error",
			role: "assistant",
			parts: []OpenAIPartIn{
				{
					Type: "tool_call",
					ID:   "call_123",
				},
			},
			wantErr: true,
		},
		{
			name: "tool_call with invalid JSON arguments",
			role: "assistant",
			parts: []OpenAIPartIn{
				{
					Type: "tool_call",
					ID:   "call_123",
					Function: &OpenAIFunctionCall{
						Name:      "tool",
						Arguments: `{invalid json}`,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "tool_result without tool_call_id",
			role: "user",
			parts: []OpenAIPartIn{
				{
					Type:   "tool_result",
					Output: "Result",
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported part type",
			role: "user",
			parts: []OpenAIPartIn{
				{Type: "unknown_type"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedRole, normalizedParts, err := normalizer.NormalizeFromOpenAI(tt.role, tt.parts)

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
