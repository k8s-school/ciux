package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTags(t *testing.T) {
	assert := assert.New(t)
	tags, err := ListTags("docker.io/library/alpine")
	assert.NoError(err)
	t.Logf("Alpine Tags: %+v", tags)

	tags, err = ListTags("gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker")
	assert.NoError(err)
	t.Logf("Fink Tags: %+v", tags)
}

func TestGetImage(t *testing.T) {
	assert := assert.New(t)
	_, ref, err := GetImage("docker.io/library/alpine:3.18.3")
	assert.NoError(err)
	assert.Equal("index.docker.io/library/alpine:3.18.3", ref.Name())
	t.Logf("Alpine ref %+v", ref)

	_, ref, err = GetImage("gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:2.7.1-104-g9cc0522")
	assert.NoError(err)
	assert.Equal("gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:2.7.1-104-g9cc0522", ref.Name())
	t.Logf("Fink ref: %+v", ref)

	_, ref, err = GetImage("gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:notexist")
	assert.Error(err)
	assert.Nil(ref)
}
