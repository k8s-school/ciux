package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require := require.New(t)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			actual, err := LastDir(tt.path)
			require.NoError(err)
			require.Equal(tt.expected, actual)
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
			name:           "invalid subdirectory",
			filePath:       "/home/user/documents/file.txt",
			subdirectory:   "/tmp",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "valid subdirectory",
			filePath:       "/cwd/documents/file.txt",
			subdirectory:   "/cwd/documents",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "empty filePath",
			filePath:       "",
			subdirectory:   "/home/user",
			expectedResult: false,
			expectedError:  fmt.Errorf("invalid argument: absFilePath=%q", ""),
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

	require := require.New(t)

	tmpSourceDir, err := os.MkdirTemp(os.TempDir(), "ciux-IsFileInSourcePathes")
	if err != nil {
		t.Errorf("Error creating temporary directory: %v", err)
	}
	t.Logf("Created temporary directory: %s", tmpSourceDir)

	tests := []struct {
		name           string
		root           string
		filePath       string
		sourcePathes   []string
		expectedResult bool
		expectedError  error
	}{

		{
			name:           "not existing subdirectory",
			root:           "",
			filePath:       "/home/user/documents/file.txt",
			sourcePathes:   []string{"/home/user"},
			expectedResult: false,
			expectedError:  fmt.Errorf("invalid argument filePath must be relative: filePath=\"/home/user/documents/file.txt\""),
		},
		{
			name:           "valid subdirectory",
			root:           tmpSourceDir,
			filePath:       "cwd/documents/file.txt",
			sourcePathes:   []string{"cwd", "cwd/documents"},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "invalid subdirectory",
			root:           tmpSourceDir,
			filePath:       "cwd/documents/file.txt",
			sourcePathes:   []string{"not-exist"},
			expectedResult: false,
			expectedError:  &fs.PathError{Op: "stat", Path: filepath.Join(tmpSourceDir, "not-exist"), Err: syscall.ENOENT},
		},
		{
			name:           "empty filePath",
			root:           tmpSourceDir,
			filePath:       "",
			sourcePathes:   []string{"source-path"},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty subdirectories list",
			root:           tmpSourceDir,
			filePath:       "/home/user/documents/file.txt",
			sourcePathes:   []string{},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "sourcePathes contains a file",
			root:           tmpSourceDir,
			filePath:       "Dockerfile",
			sourcePathes:   []string{"tmp", "Dockerfile"},
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {

		for _, path := range tt.sourcePathes {
			if path == "Dockerfile" {
				file := filepath.Join(tmpSourceDir, path)
				_, err := os.Create(file)
				if err != nil {
					t.Errorf("Error creating Dockerfile: %v", err)
				}
			} else if path != "not-exist" && path[0] != '/' {
				dir := filepath.Join(tmpSourceDir, path)
				err := os.Mkdir(dir, 0755)
				if err != nil {
					t.Errorf("Error creating directory %s: %v", path, err)
				}
			}
		}

		t.Run(tt.name, func(t *testing.T) {

			actualResult, actualError := IsFileInSourcePathes(tt.root, tt.filePath, tt.sourcePathes)
			require.Equal(tt.expectedError, actualError)

			require.Equal(tt.expectedResult, actualResult)

		})
	}
}
