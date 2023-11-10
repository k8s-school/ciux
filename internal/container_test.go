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

func TestDescImage(t *testing.T) {
	assert := assert.New(t)
	_, ref, err := DescImage("docker.io/library/alpine:3.18.3")
	assert.NoError(err)
	assert.Equal("index.docker.io/library/alpine:3.18.3", ref.Name())
	t.Logf("Alpine ref %+v", ref)

	_, ref, err = DescImage("docker.io/library/alpine:3.18.3:notexist")
	assert.Error(err)
	assert.Nil(ref)
}
