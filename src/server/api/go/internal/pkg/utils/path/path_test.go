package path

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorType   error
	}{
		// Valid paths (directories only, no files)
		{
			name:        "valid simple directory",
			path:        "documents",
			expectError: false,
		},
		{
			name:        "valid nested directory",
			path:        "folder/subfolder",
			expectError: false,
		},
		{
			name:        "valid directory with special characters",
			path:        "folder/subfolder_v2.1",
			expectError: false,
		},
		{
			name:        "valid directory with numbers",
			path:        "2023/12/25",
			expectError: false,
		},
		{
			name:        "valid root directory path",
			path:        "/",
			expectError: false,
		},
		{
			name:        "valid path with leading slash",
			path:        "/folder/subfolder",
			expectError: false,
		},
		{
			name:        "valid path with trailing slash",
			path:        "folder/subfolder/",
			expectError: false,
		},
		{
			name:        "valid directory with special characters",
			path:        "folder<name>",
			expectError: false,
		},
		{
			name:        "valid directory with reserved name",
			path:        "CON",
			expectError: false,
		},
		{
			name:        "valid path with file extension (now allowed)",
			path:        "document.txt",
			expectError: false,
		},
		{
			name:        "valid nested path with file extension (now allowed)",
			path:        "folder/document.pdf",
			expectError: false,
		},

		// Empty path
		{
			name:        "empty path",
			path:        "",
			expectError: true,
			errorType:   ErrEmptyPath,
		},

		// Directory traversal
		{
			name:        "directory traversal with ..",
			path:        "../secret.txt",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "directory traversal in middle",
			path:        "folder/../secret.txt",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "multiple dots",
			path:        ".../file.txt",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with null byte",
			path:        "file\x00name.txt",
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "path with only dots",
			path:        "...",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with multiple consecutive dots",
			path:        "....",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with dots at end",
			path:        "folder/...",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with dots at beginning",
			path:        ".../folder",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with single dot",
			path:        ".",
			expectError: false,
		},
		{
			name:        "path with single dot in directory",
			path:        "folder/.",
			expectError: false,
		},
		{
			name:        "path with whitespace only",
			path:        "   ",
			expectError: false,
		},
		{
			name:        "path with tabs",
			path:        "\t",
			expectError: false,
		},
		{
			name:        "path with newlines",
			path:        "folder\n/file",
			expectError: false,
		},
		{
			name:        "path with carriage return",
			path:        "folder\r/file",
			expectError: false,
		},
		{
			name:        "very long path",
			path:        "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/1/2/3/4/5/6/7/8/9/0",
			expectError: false,
		},
		{
			name:        "path with unicode characters",
			path:        "文件夹/文件.txt",
			expectError: false,
		},
		{
			name:        "path with emoji",
			path:        "📁/📄.txt",
			expectError: false,
		},
		{
			name:        "path with null byte in middle",
			path:        "folder/file\x00name.txt",
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "path with null byte at start",
			path:        "\x00file.txt",
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "path with null byte at end",
			path:        "file.txt\x00",
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "path with only one dot",
			path:        ".",
			expectError: false,
		},
		{
			name:        "path with two dots",
			path:        "..",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with mixed dots and other chars",
			path:        "file...name",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "path with dots at start of filename",
			path:        "folder/.hidden",
			expectError: false,
		},
		{
			name:        "path with empty parts",
			path:        "//folder//file",
			expectError: false,
		},
		{
			name:        "path with trailing slash",
			path:        "folder/file/",
			expectError: false,
		},
		{
			name:        "path with leading slash",
			path:        "/folder/file",
			expectError: false,
		},
		{
			name:        "path with both leading and trailing slashes",
			path:        "/folder/file/",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean directory path",
			input:    "documents",
			expected: "documents",
		},
		{
			name:     "path with leading slash",
			input:    "/documents",
			expected: "documents",
		},
		{
			name:     "path with leading dots",
			input:    "./documents",
			expected: "documents",
		},
		{
			name:     "path with null byte",
			input:    "file\x00name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "path starting with dots",
			input:    ".../secret",
			expected: "file_.../secret",
		},
		{
			name:     "path with special characters",
			input:    "file<name>",
			expected: "file<name>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDirectoriesFromPaths(t *testing.T) {
	tests := []struct {
		name       string
		parentPath string
		filePaths  []string
		expected   []string
	}{
		{
			name:       "root path with nested directories",
			parentPath: "/",
			filePaths: []string{
				"/documents/file1.txt",
				"/documents/file2.pdf",
				"/images/photo1.jpg",
				"/images/photo2.png",
				"/code/script.py",
			},
			expected: []string{"documents", "images", "code"},
		},
		{
			name:       "nested parent path",
			parentPath: "/documents",
			filePaths: []string{
				"/documents/work/project1.txt",
				"/documents/work/project2.txt",
				"/documents/personal/note1.txt",
				"/documents/personal/note2.txt",
				"/images/photo.jpg", // This should be ignored
			},
			expected: []string{"work", "personal"},
		},
		{
			name:       "parent path with trailing slash",
			parentPath: "/documents/",
			filePaths: []string{
				"/documents/work/project1.txt",
				"/documents/personal/note1.txt",
			},
			expected: []string{"work", "personal"},
		},
		{
			name:       "no matching paths",
			parentPath: "/nonexistent",
			filePaths: []string{
				"/documents/file1.txt",
				"/images/photo.jpg",
			},
			expected: []string{},
		},
		{
			name:       "files directly in parent path",
			parentPath: "/documents",
			filePaths: []string{
				"/documents/file1.txt",
				"/documents/file2.pdf",
			},
			expected: []string{"file1.txt", "file2.pdf"},
		},
		{
			name:       "empty parent path defaults to root",
			parentPath: "",
			filePaths: []string{
				"/documents/file1.txt",
				"/images/photo.jpg",
			},
			expected: []string{"documents", "images"},
		},
		{
			name:       "single directory",
			parentPath: "/",
			filePaths: []string{
				"/single/file.txt",
			},
			expected: []string{"single"},
		},
		{
			name:       "duplicate directories should be unique",
			parentPath: "/",
			filePaths: []string{
				"/documents/file1.txt",
				"/documents/file2.txt",
				"/images/photo1.jpg",
				"/images/photo2.jpg",
			},
			expected: []string{"documents", "images"},
		},
		{
			name:       "paths without leading slash should be normalized",
			parentPath: "/",
			filePaths: []string{
				"webp/image1.webp",
				"webp/image2.webp",
				"documents/file1.txt",
				"images/photo.jpg",
			},
			expected: []string{"webp", "documents", "images"},
		},
		{
			name:       "parent path without leading slash should be normalized",
			parentPath: "documents",
			filePaths: []string{
				"/documents/work/project1.txt",
				"/documents/personal/note1.txt",
				"/images/photo.jpg",
			},
			expected: []string{"work", "personal"},
		},
		{
			name:       "paths with extra spaces should be normalized",
			parentPath: " / ",
			filePaths: []string{
				" /webp/image1.webp ",
				"/documents/file1.txt",
				" /images/photo.jpg ",
			},
			expected: []string{"webp", "documents", "images"},
		},
		{
			name:       "root path query with single directory",
			parentPath: "/",
			filePaths: []string{
				"/",
				"/webp",
			},
			expected: []string{"webp"},
		},
		{
			name:       "empty file paths list",
			parentPath: "/",
			filePaths:  []string{},
			expected:   []string{},
		},
		{
			name:       "file paths with empty strings",
			parentPath: "/",
			filePaths: []string{
				"",
				"   ",
				"/documents/file1.txt",
			},
			expected: []string{"documents"},
		},
		{
			name:       "file paths with only root",
			parentPath: "/",
			filePaths: []string{
				"/",
			},
			expected: []string{},
		},
		{
			name:       "file paths with unicode characters",
			parentPath: "/",
			filePaths: []string{
				"/文件夹/文件1.txt",
				"/文件夹/文件2.txt",
				"/📁/📄.txt",
			},
			expected: []string{"文件夹", "📁"},
		},
		{
			name:       "file paths with special characters",
			parentPath: "/",
			filePaths: []string{
				"/folder-name/file1.txt",
				"/folder_name/file2.txt",
				"/folder.name/file3.txt",
			},
			expected: []string{"folder-name", "folder_name", "folder.name"},
		},
		{
			name:       "file paths with very long names",
			parentPath: "/",
			filePaths: []string{
				"/very-long-directory-name-that-exceeds-normal-limits/file1.txt",
				"/another-very-long-directory-name/file2.txt",
			},
			expected: []string{"very-long-directory-name-that-exceeds-normal-limits", "another-very-long-directory-name"},
		},
		{
			name:       "file paths with numbers",
			parentPath: "/",
			filePaths: []string{
				"/2023/file1.txt",
				"/2024/file2.txt",
				"/v1.0/file3.txt",
			},
			expected: []string{"2023", "2024", "v1.0"},
		},
		{
			name:       "file paths with mixed case",
			parentPath: "/",
			filePaths: []string{
				"/Documents/file1.txt",
				"/DOCUMENTS/file2.txt",
				"/documents/file3.txt",
			},
			expected: []string{"Documents", "DOCUMENTS", "documents"},
		},
		{
			name:       "parent path with special characters",
			parentPath: "/special-folder/",
			filePaths: []string{
				"/special-folder/sub-folder/file1.txt",
				"/special-folder/sub_folder/file2.txt",
				"/special-folder/sub.folder/file3.txt",
			},
			expected: []string{"sub-folder", "sub_folder", "sub.folder"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDirectoriesFromPaths(tt.parentPath, tt.filePaths)

			// Sort both slices for comparison since order doesn't matter
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitFilePath(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedPath string
		expectedFile string
	}{
		{
			name:         "simple file in root",
			input:        "/report.pdf",
			expectedPath: "/",
			expectedFile: "report.pdf",
		},
		{
			name:         "file in nested directory",
			input:        "/documents/report.pdf",
			expectedPath: "/documents/",
			expectedFile: "report.pdf",
		},
		{
			name:         "file without leading slash",
			input:        "report.pdf",
			expectedPath: "/",
			expectedFile: "report.pdf",
		},
		{
			name:         "directory path only",
			input:        "/documents/",
			expectedPath: "/documents/",
			expectedFile: "",
		},
		{
			name:         "root directory",
			input:        "/",
			expectedPath: "/",
			expectedFile: "",
		},
		{
			name:         "empty string",
			input:        "",
			expectedPath: "/",
			expectedFile: "",
		},
		{
			name:         "file with spaces",
			input:        " report.pdf ",
			expectedPath: "/",
			expectedFile: "report.pdf",
		},
		{
			name:         "nested path with spaces",
			input:        " /documents/report.pdf ",
			expectedPath: "/documents/",
			expectedFile: "report.pdf",
		},
		{
			name:         "file with special characters",
			input:        "/documents/report_v2.1.pdf",
			expectedPath: "/documents/",
			expectedFile: "report_v2.1.pdf",
		},
		{
			name:         "deeply nested path",
			input:        "/folder1/folder2/folder3/file.txt",
			expectedPath: "/folder1/folder2/folder3/",
			expectedFile: "file.txt",
		},
		{
			name:         "file with no extension",
			input:        "/documents/README",
			expectedPath: "/documents/",
			expectedFile: "README",
		},
		{
			name:         "single character filename",
			input:        "/a",
			expectedPath: "/",
			expectedFile: "a",
		},
		{
			name:         "filename with dots",
			input:        "/documents/file.name.txt",
			expectedPath: "/documents/",
			expectedFile: "file.name.txt",
		},
		{
			name:         "path with multiple slashes",
			input:        "//documents//report.pdf",
			expectedPath: "//documents//",
			expectedFile: "report.pdf",
		},
		{
			name:         "filename only",
			input:        "filename.txt",
			expectedPath: "/",
			expectedFile: "filename.txt",
		},
		{
			name:         "path with only slashes",
			input:        "///",
			expectedPath: "///",
			expectedFile: "",
		},
		{
			name:         "path ending with slash",
			input:        "/documents/",
			expectedPath: "/documents/",
			expectedFile: "",
		},
		{
			name:         "path with trailing spaces",
			input:        "/documents/file.txt   ",
			expectedPath: "/documents/",
			expectedFile: "file.txt",
		},
		{
			name:         "path with leading spaces",
			input:        "   /documents/file.txt",
			expectedPath: "/documents/",
			expectedFile: "file.txt",
		},
		{
			name:         "path with no slash found",
			input:        "filename",
			expectedPath: "/",
			expectedFile: "filename",
		},
		{
			name:         "path with only filename and spaces",
			input:        "  filename  ",
			expectedPath: "/",
			expectedFile: "filename",
		},
		{
			name:         "path with only slash",
			input:        "/",
			expectedPath: "/",
			expectedFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, filename := SplitFilePath(tt.input)
			assert.Equal(t, tt.expectedPath, path)
			assert.Equal(t, tt.expectedFile, filename)
		})
	}
}
