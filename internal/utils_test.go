package internal

import (
	"fmt"
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
func TestIsPathInSubdirectories(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		subdirectories []string
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "valid absolute subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectories: []string{"/home/user", "/home/toto"},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid absolute subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectories: []string{"/tmp"},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "valid subdirectory",
			filePath:       "cwd/documents/file.txt",
			subdirectories: []string{"cwd/documents"},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid absolute subdirectory",
			filePath:       "cwd/documents/file.txt",
			subdirectories: []string{"tmp"},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty filePath",
			filePath:       "",
			subdirectories: []string{"/home/user"},
			expectedResult: false,
			expectedError:  fmt.Errorf("invalid arguments: filePath=%q, subdirectory=%q", "", "/home/user"),
		},
		{
			name:           "empty subdirectories list",
			filePath:       "/home/user/documents/file.txt",
			subdirectories: []string{},
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualResult, actualError := IsPathInSubdirectories(tt.filePath, tt.subdirectories)
			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)
		})
	}
}
