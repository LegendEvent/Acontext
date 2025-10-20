package fileparser

import (
	"testing"
)

func TestFileParser(t *testing.T) {
	parser := NewFileParser()

	tests := []struct {
		name     string
		filename string
		mimeType string
		content  []byte
		expected string
	}{
		{
			name:     "JSON file",
			filename: "test.json",
			mimeType: "application/json",
			content:  []byte(`{"name": "test", "value": 123}`),
			expected: "json",
		},
		{
			name:     "CSV file",
			filename: "test.csv",
			mimeType: "text/csv",
			content:  []byte("name,age\nJohn,25\nJane,30"),
			expected: "csv",
		},
		{
			name:     "Go code file",
			filename: "test.go",
			mimeType: "text/x-go",
			content:  []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"),
			expected: "code",
		},
		{
			name:     "Markdown file",
			filename: "test.md",
			mimeType: "text/markdown",
			content:  []byte("# Test\n\nThis is a test markdown file."),
			expected: "text",
		},
		{
			name:     "Plain text file",
			filename: "test.txt",
			mimeType: "text/plain",
			content:  []byte("This is plain text content."),
			expected: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseFile(tt.filename, tt.mimeType, tt.content)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			if result.Type != tt.expected {
				t.Errorf("ParseFile() type = %v, want %v", result.Type, tt.expected)
			}

			if result.Raw != string(tt.content) {
				t.Errorf("ParseFile() raw content = %v, want %v", result.Raw, string(tt.content))
			}
		})
	}
}

func TestJSONParser(t *testing.T) {
	parser := &JSONParser{}

	// Test valid JSON
	content := []byte(`{"name": "test", "value": 123}`)
	result, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("JSONParser.Parse() error = %v", err)
	}

	if result.Type != "json" {
		t.Errorf("JSONParser.Parse() type = %v, want json", result.Type)
	}

	// Test invalid JSON
	invalidContent := []byte(`{"name": "test", "value": 123`) // Missing closing brace
	_, err = parser.Parse(invalidContent)
	if err == nil {
		t.Error("JSONParser.Parse() should return error for invalid JSON")
	}
}

func TestCSVParser(t *testing.T) {
	parser := &CSVParser{}

	content := []byte("name,age\nJohn,25\nJane,30")
	result, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("CSVParser.Parse() error = %v", err)
	}

	if result.Type != "csv" {
		t.Errorf("CSVParser.Parse() type = %v, want csv", result.Type)
	}

	if result.Raw != string(content) {
		t.Errorf("CSVParser.Parse() raw = %v, want %v", result.Raw, string(content))
	}
}

func TestUnsupportedFileType(t *testing.T) {
	parser := NewFileParser()

	// Test image file (should not be supported)
	_, err := parser.ParseFile("image.png", "image/png", []byte{})
	if err == nil {
		t.Error("ParseFile() should return error for unsupported file type")
	}

	// Test binary file
	_, err = parser.ParseFile("binary.exe", "application/octet-stream", []byte{})
	if err == nil {
		t.Error("ParseFile() should return error for unsupported file type")
	}
}
