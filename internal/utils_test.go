package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLastDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			path:     "/home/user/documents/file.txt",
			expected: "file.txt",
		},
		{
			name:     "path with trailing slash",
			path:     "/home/user/documents/",
			expected: "documents",
		},
		{
			name:     "path with multiple separators",
			path:     "/home/user/documents//file.txt",
			expected: "file.txt",
		},
		{
			name:     "empty path",
			path:     "",
			expected: ".",
		},
	}
	assert := assert.New(t)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			actual, err := LastDir(tt.path)
			assert.NoError(err)
			assert.Equal(tt.expected, actual)
		})
	}
}

func TestIsPathInSubdirectory(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		subdirectory   string
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "valid absolute subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectory:   "/home/user",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid absolute subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectory:   "/tmp",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "valid subdirectory",
			filePath:       "cwd/documents/file.txt",
			subdirectory:   "cwd/documents",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid absolute subdirectory",
			filePath:       "cwd/documents/file.txt",
			subdirectory:   "tmp",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty filePath",
			filePath:       "",
			subdirectory:   "/home/user",
			expectedResult: false,
			expectedError:  fmt.Errorf("invalid arguments: filePath=%q, subdirectory=%q", "", "/home/user"),
		},
		{
			name:           "empty subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectory:   "",
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualResult, actualError := IsPathInSubdirectory(tt.filePath, tt.subdirectory)
			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)
		})
	}
}
func TestIsFileInSourcePathes(t *testing.T) {

	tmpSourceDir, err := os.MkdirTemp(os.TempDir(), "ciux-IsFileInSourcePathes")
	if err != nil {
		t.Errorf("Error creating temporary directory: %v", err)
	}

	// In ciux, file comparison is made in the root of the source directory
	err = os.Chdir(tmpSourceDir)
	if err != nil {
		t.Errorf("Error changing directory to %s: %v", tmpSourceDir, err)
	}

	tests := []struct {
		name           string
		filePath       string
		sourcePathes   []string
		expectedResult bool
		expectedError  error
	}{

		{
			name:           "absolute subdirectory",
			filePath:       "/home/user/documents/file.txt",
			sourcePathes:   []string{"/home/user", "/home/toto"},
			expectedResult: false,
			expectedError:  errors.New("invalid source path, must be relative: \"/home/user\""),
		},
		{
			name:           "valid subdirectory",
			filePath:       "cwd/documents/file.txt",
			sourcePathes:   []string{"cwd", "cwd/documents"},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid absolute subdirectory",
			filePath:       "cwd/documents/file.txt",
			sourcePathes:   []string{"invalid-directory"},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty filePath",
			filePath:       "",
			sourcePathes:   []string{"source-path"},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty subdirectories list",
			filePath:       "/home/user/documents/file.txt",
			sourcePathes:   []string{},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "sourcePathes contains a file",
			filePath:       "Dockerfile",
			sourcePathes:   []string{"tmp", "Dockerfile"},
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, path := range tt.sourcePathes {
				if path == "Dockerfile" {
					file := filepath.Join(tmpSourceDir, path)
					_, err := os.Create(file)
					if err != nil {
						t.Errorf("Error creating Dockerfile: %v", err)
					}
				} else if path != "invalid-directory" && path[0] != '/' {
					dir := filepath.Join(tmpSourceDir, path)
					err := os.Mkdir(dir, 0755)
					if err != nil {
						t.Errorf("Error creating directory %s: %v", path, err)
					}
				}
			}

			actualResult, actualError := IsFileInSourcePathes(tt.filePath, tt.sourcePathes)
			assert.Equal(t, tt.expectedResult, actualResult)
			fmt.Printf("actualError: %v\n", actualError)

			if tt.expectedResult == false && tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, actualError)
			}
		})
	}
}
