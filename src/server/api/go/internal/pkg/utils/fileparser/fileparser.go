package fileparser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"
)

// FileContent represents the parsed content of a file
type FileContent struct {
	Type string `json:"type"` // "text", "json", "csv", "code"
	Raw  string `json:"raw"`  // Raw text content
}

// Parser interface for different file types
type Parser interface {
	CanParse(filename string, mimeType string) bool
	Parse(content []byte) (*FileContent, error)
}

// TextParser handles plain text files
type TextParser struct{}

func (p *TextParser) CanParse(filename string, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExts := []string{".txt", ".md", ".markdown", ".log", ".yml", ".yaml", ".xml", ".html", ".htm"}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	// Check MIME types
	textMimes := []string{"text/", "application/xml", "application/x-yaml"}
	for _, mime := range textMimes {
		if strings.HasPrefix(mimeType, mime) {
			return true
		}
	}

	return false
}

func (p *TextParser) Parse(content []byte) (*FileContent, error) {
	return &FileContent{
		Type: "text",
		Raw:  string(content),
	}, nil
}

// JSONParser handles JSON files
type JSONParser struct{}

func (p *JSONParser) CanParse(filename string, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".json" {
		return true
	}
	return strings.HasPrefix(mimeType, "application/json")
}

func (p *JSONParser) Parse(content []byte) (*FileContent, error) {
	// Validate that it's valid JSON using sonic
	var jsonData interface{}
	if err := sonic.Unmarshal(content, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &FileContent{
		Type: "json",
		Raw:  string(content),
	}, nil
}

// CSVParser handles CSV files
type CSVParser struct{}

func (p *CSVParser) CanParse(filename string, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return true
	}
	return strings.HasPrefix(mimeType, "text/csv") || mimeType == "application/csv"
}

func (p *CSVParser) Parse(content []byte) (*FileContent, error) {
	// Validate that it's valid CSV
	reader := csv.NewReader(bytes.NewReader(content))
	_, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return &FileContent{
		Type: "csv",
		Raw:  string(content),
	}, nil
}

// CodeParser handles code files
type CodeParser struct{}

func (p *CodeParser) CanParse(filename string, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	codeExts := []string{
		".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".java", ".cpp", ".c", ".h", ".hpp",
		".cs", ".php", ".rb", ".rs", ".swift", ".kt", ".scala", ".sh", ".bash", ".zsh",
		".sql", ".r", ".m", ".pl", ".lua", ".vim", ".dockerfile", ".makefile", ".cmake",
		".proto", ".thrift", ".graphql", ".gql", ".vue", ".svelte", ".astro",
	}

	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}

	// Check for common code MIME types
	codeMimes := []string{
		"text/x-go", "text/x-python", "text/javascript", "text/typescript",
		"application/javascript", "application/typescript", "text/x-java-source",
		"text/x-c", "text/x-c++", "text/x-csharp", "text/x-php", "text/x-ruby",
		"text/x-rust", "text/x-swift", "text/x-kotlin", "text/x-scala",
		"text/x-shellscript", "application/sql", "text/x-r", "text/x-perl",
		"text/x-lua", "text/x-vim", "text/x-dockerfile", "text/x-makefile",
		"text/x-protobuf", "text/x-thrift", "application/graphql",
	}

	for _, mime := range codeMimes {
		if strings.HasPrefix(mimeType, mime) {
			return true
		}
	}

	return false
}

func (p *CodeParser) Parse(content []byte) (*FileContent, error) {
	return &FileContent{
		Type: "code",
		Raw:  string(content),
	}, nil
}

// FileParser manages all parsers
type FileParser struct {
	parsers []Parser
}

// NewFileParser creates a new file parser with all available parsers
func NewFileParser() *FileParser {
	return &FileParser{
		parsers: []Parser{
			&JSONParser{},
			&CSVParser{},
			&CodeParser{},
			&TextParser{}, // Text parser should be last as it's the fallback
		},
	}
}

// CanParseFile checks if a file can be parsed based on filename and MIME type
func (fp *FileParser) CanParseFile(filename string, mimeType string) bool {
	for _, parser := range fp.parsers {
		if parser.CanParse(filename, mimeType) {
			return true
		}
	}
	return false
}

// ParseFile attempts to parse file content based on filename and MIME type
func (fp *FileParser) ParseFile(filename string, mimeType string, content []byte) (*FileContent, error) {
	// Try each parser in order
	for _, parser := range fp.parsers {
		if parser.CanParse(filename, mimeType) {
			return parser.Parse(content)
		}
	}

	// If no parser matches (e.g., images, binary files), return nil
	// Only text-based files should be parsed
	return nil, fmt.Errorf("unsupported file type: %s (mime: %s)", filename, mimeType)
}

// ParseFileFromReader parses file content from an io.Reader
func (fp *FileParser) ParseFileFromReader(filename string, mimeType string, reader io.Reader) (*FileContent, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return fp.ParseFile(filename, mimeType, content)
}
