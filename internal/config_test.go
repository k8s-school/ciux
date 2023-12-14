package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)
	configPath, err := os.Getwd()
	assert.NoError(err)

	c, err := NewConfig(configPath)
	assert.NoError(err)
	assert.Equal("test-registry.io", c.Registry)

	var labelSet = labels.Set{
		"key1": "value1",
		"key2": "value2",
	}

	expertedDep := DepConfig{
		Url:    "file:///tmp/ciux-dep-test",
		Clone:  true,
		Pull:   true,
		Labels: labelSet,
	}

	assert.Equal(expertedDep, c.Dependencies[0])
}
