package internal

import (
	"testing"

	require "github.com/stretchr/testify/assert"
)

func TestListTags(t *testing.T) {
	assert := require.New(t)
	tags, err := ListTags("docker.io/library/alpine")
	assert.NoError(err)
	t.Logf("Alpine Tags: %+v", tags)

	tags, err = ListTags("gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker")
	assert.NoError(err)
	t.Logf("Fink Tags: %+v", tags)
}

func TestDescImage(t *testing.T) {
	require := require.New(t)
	_, ref, err := DescImage("docker.io/library/alpine:3.18.3")
	require.NoError(err)
	require.Equal("index.docker.io/library/alpine:3.18.3", ref.Name())
	t.Logf("Alpine ref %+v", ref)

	_, ref, err = DescImage("docker.io/library/alpine:3.18.3:notexist")
	require.Error(err)
	require.Nil(ref)
}
func TestGetImgEnVarPrefix(t *testing.T) {
	assert := require.New(t)

	image := "docker.io/library/alpine:3.18.3"
	expectedPrefix := "LIBRARY_ALPINE"

	prefix, err := GetImageEnVarPrefix(image)
	assert.NoError(err)
	assert.Equal(expectedPrefix, prefix)

	image = "gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:latest"
	expectedPrefix = "ASTROLABSOFTWARE_FINK_FINK_BROKER"

	prefix, err = GetImageEnVarPrefix(image)
	assert.NoError(err)
	assert.Equal(expectedPrefix, prefix)
}
