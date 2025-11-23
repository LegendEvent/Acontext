package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTemplatesConfig(t *testing.T) {
	config, err := LoadTemplatesConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Repo)
	assert.NotEmpty(t, config.Templates)
}

func TestGetLanguages(t *testing.T) {
	languages := GetLanguages()
	assert.NotEmpty(t, languages)
	assert.Contains(t, languages, "python")
	assert.Contains(t, languages, "typescript")
}

func TestNeedsTemplateDiscovery(t *testing.T) {
	// Test with languages that should need discovery (no presets in YAML)
	needsDiscovery, err := NeedsTemplateDiscovery("python")
	assert.NoError(t, err)
	assert.True(t, needsDiscovery, "python should need template discovery")

	needsDiscovery, err = NeedsTemplateDiscovery("typescript")
	assert.NoError(t, err)
	assert.True(t, needsDiscovery, "typescript should need template discovery")
}

func TestFormatTemplateName(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		templateName string
		expected     string
	}{
		{
			name:         "simple name",
			language:     "python",
			templateName: "openai",
			expected:     "Openai",
		},
		{
			name:         "hyphenated name",
			language:     "typescript",
			templateName: "vercel-ai",
			expected:     "Vercel Ai",
		},
		{
			name:         "underscore name",
			language:     "python",
			templateName: "custom_template",
			expected:     "Custom Template",
		},
		{
			name:         "mixed separators",
			language:     "typescript",
			templateName: "my-custom_template",
			expected:     "My Custom Template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTemplateName(tt.language, tt.templateName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
