package converter

import (
	"testing"

	"github.com/memodb-io/Acontext/internal/modules/service"
	"github.com/stretchr/testify/assert"
)

func TestGetNormalizer(t *testing.T) {
	tests := []struct {
		name    string
		format  MessageFormat
		wantErr bool
	}{
		{
			name:    "None format returns NoOpNormalizer",
			format:  FormatAcontext,
			wantErr: false,
		},
		{
			name:    "OpenAI format returns OpenAINormalizer",
			format:  FormatOpenAI,
			wantErr: false,
		},
		{
			name:    "Anthropic format returns AnthropicNormalizer",
			format:  FormatAnthropic,
			wantErr: false,
		},
		{
			name:    "Invalid format returns error",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer, err := GetNormalizer(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, normalizer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, normalizer)
			}
		})
	}
}

func TestNoOpNormalizer_Normalize(t *testing.T) {
	normalizer := &NoOpNormalizer{}

	role := "user"
	parts := []service.PartIn{
		{Type: "text", Text: "Hello"},
	}

	normalizedRole, normalizedParts, err := normalizer.Normalize(role, parts)

	assert.NoError(t, err)
	assert.Equal(t, role, normalizedRole)
	assert.Equal(t, parts, normalizedParts)
}
