package internal

import (
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
