package internal

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)
	c, err := NewConfig("")
	assert.NoError(err)
	assert.Equal("test-registry.io", viper.AllSettings()["registry"])
	assert.Equal("test-registry.io", c.Registry)

	dep := Dependency{
		Url:   "file:///tmp/ciux-dep-test",
		Clone: true,
		Pull:  true,
	}

	assert.Equal(dep, c.Dependencies[0])
}
