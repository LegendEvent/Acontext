package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeProjectNameForPackage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		format   string
		expected string
	}{
		{
			name:     "Python: simple name",
			input:    "MyProject",
			format:   "python",
			expected: "myproject",
		},
		{
			name:     "Python: with hyphens",
			input:    "my-acontext-app",
			format:   "python",
			expected: "my_acontext_app",
		},
		{
			name:     "Python: with spaces",
			input:    "My Acontext App",
			format:   "python",
			expected: "my_acontext_app",
		},
		{
			name:     "npm: simple name",
			input:    "MyProject",
			format:   "npm",
			expected: "myproject",
		},
		{
			name:     "npm: with underscores",
			input:    "my_acontext_app",
			format:   "npm",
			expected: "my-acontext-app",
		},
		{
			name:     "npm: with spaces",
			input:    "My Acontext App",
			format:   "npm",
			expected: "my-acontext-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeProjectNameForPackage(tt.input, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReplacePyProjectName(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-pyproject-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Create a test pyproject.toml
	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	pyprojectContent := `[project]
name = "acontext-examples"
version = "0.1.0"
description = "Add your description here"
requires-python = ">=3.13"
dependencies = [
    "acontext",
    "python-dotenv",
]
`
	err = os.WriteFile(pyprojectPath, []byte(pyprojectContent), 0644)
	require.NoError(t, err)

	// Replace name
	err = replacePyProjectName(pyprojectPath, "My-Acontext-App")
	require.NoError(t, err)

	// Read and verify
	data, err := os.ReadFile(pyprojectPath)
	require.NoError(t, err)

	// Check that the name was updated (TOML library may use single or double quotes)
	assert.Contains(t, string(data), "name =")
	assert.Contains(t, string(data), "my_acontext_app")
	assert.NotContains(t, string(data), "acontext-examples")
}

func TestReplacePackageJsonName(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-package-json-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Create a test package.json
	packageJsonPath := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
  "name": "acontext-examples",
  "version": "0.1.0",
  "description": "Add your description here"
}
`
	err = os.WriteFile(packageJsonPath, []byte(packageJsonContent), 0644)
	require.NoError(t, err)

	// Replace name
	err = replacePackageJsonName(packageJsonPath, "My-Acontext-App")
	require.NoError(t, err)

	// Read and verify
	data, err := os.ReadFile(packageJsonPath)
	require.NoError(t, err)

	// Check that the name was updated
	assert.Contains(t, string(data), `"name": "my-acontext-app"`)
	assert.NotContains(t, string(data), "acontext-examples")
}

func TestReplaceTemplateVars(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-template-vars-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Create test files
	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	pyprojectContent := `[project]
name = "acontext-examples"
version = "0.1.0"
`
	err = os.WriteFile(pyprojectPath, []byte(pyprojectContent), 0644)
	require.NoError(t, err)

	packageJsonPath := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
  "name": "acontext-examples",
  "version": "0.1.0"
}
`
	err = os.WriteFile(packageJsonPath, []byte(packageJsonContent), 0644)
	require.NoError(t, err)

	// Replace template variables
	vars := map[string]string{
		"project_name": "My-New-Project",
	}
	err = replaceTemplateVars(tempDir, vars)
	require.NoError(t, err)

	// Verify pyproject.toml
	pyprojectData, err := os.ReadFile(pyprojectPath)
	require.NoError(t, err)
	assert.Contains(t, string(pyprojectData), "name =")
	assert.Contains(t, string(pyprojectData), "my_new_project")

	// Verify package.json
	packageJsonData, err := os.ReadFile(packageJsonPath)
	require.NoError(t, err)
	assert.Contains(t, string(packageJsonData), `"name": "my-new-project"`)
}
