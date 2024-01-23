package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

func TestNewConfig(t *testing.T) {
	require := require.New(t)
	configPath, err := os.Getwd()
	require.NoError(err)

	c, err := NewConfig(configPath)
	require.NoError(err)
	require.Equal("test-registry.io", c.Registry)

	var labelSet = labels.Set{
		"key1": "value1",
		"key2": "value2",
	}

	expectedDep := DepConfig{
		Url:    "file:///tmp/ciux-dep-test",
		Clone:  true,
		Pull:   true,
		Labels: labelSet,
	}

	require.Equal([]string{"rootfs", "homefs"}, c.SourcePathes)
	require.Equal(expectedDep, c.Dependencies[0])
}
