package internal

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	assert := assert.New(t)
	ReadConfig()
	assert.Equal("test-registry.io", viper.AllSettings()["registry"])
}

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	ReadConfig()
	c := GetConfig()
	t.Logf("Config: %+v", c)
	assert.Equal("test-registry.io", c.Registry)

	dep := Dependency{
		Url:   "http://github.com/test",
		Clone: true,
		Pull:  true,
	}

	assert.Equal(dep, c.Dependencies[0])

}
