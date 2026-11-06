package internal

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

func randomImage(t *testing.T) v1.Image {
	img, err := random.Image(64, 2)
	require.NoError(t, err)
	return img
}

// builtOn returns an image made of base plus one extra layer, as a
// Dockerfile "FROM base" followed by one instruction would produce
func builtOn(t *testing.T, base v1.Image) v1.Image {
	layer, err := random.Layer(64, "application/vnd.docker.image.rootfs.diff.tar.gzip")
	require.NoError(t, err)
	img, err := mutate.AppendLayers(base, layer)
	require.NoError(t, err)
	return img
}

func TestIsBuiltFrom(t *testing.T) {
	base := randomImage(t)
	img := builtOn(t, base)

	tests := []struct {
		name     string
		img      v1.Image
		base     v1.Image
		expected bool
	}{
		{"built on base", img, base, true},
		{"built on another base", img, randomImage(t), false},
		{"image is the base itself", base, base, true},
		{"base has more layers than image", base, img, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBuiltFrom(tt.img, tt.base)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

// pushImage pushes img to the test registry and returns its URL
func pushImage(t *testing.T, host string, repo string, img v1.Image) string {
	url := fmt.Sprintf("%s/%s:v1", host, repo)
	ref, err := name.ParseReference(url)
	require.NoError(t, err)
	require.NoError(t, remote.Write(ref, img))
	return url
}

func TestIsBuiltFromAny(t *testing.T) {
	server := httptest.NewServer(registry.New())
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")

	scienceBase := randomImage(t)
	urlNoscience := pushImage(t, host, "deps-noscience", randomImage(t))
	urlScience := pushImage(t, host, "deps-science", scienceBase)
	urlImage := pushImage(t, host, "broker", builtOn(t, scienceBase))
	urlMissing := fmt.Sprintf("%s/deps-missing:v1", host)

	tests := []struct {
		name      string
		bases     []string
		expected  bool
		expectErr bool
	}{
		{"built on one of the bases", []string{urlNoscience, urlScience}, true, false},
		{"built on none of the bases", []string{urlNoscience}, false, false},
		{"no base declared", []string{}, true, false},
		{"base missing from the registry", []string{urlMissing}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBuiltFromAny(urlImage, tt.bases)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestBuildBaseImages(t *testing.T) {
	config := ProjConfig{
		Dependencies: []DepConfig{
			{Image: "registry/deps-noscience:v2", Labels: labels.Set{"build": "noscience"}},
			{Image: "registry/deps-science:v2", Labels: labels.Set{"build": "science"}},
			{Image: "registry/deps-all:v2", Labels: labels.Set{"build": ""}},
			{Image: "registry/hadoop:v1", Labels: labels.Set{"itest": "true"}},
			{Url: "https://github.com/org/repo", Labels: labels.Set{"build": "true"}},
			{Package: "github.com/org/tool@v1", Labels: labels.Set{"build": "true"}},
		},
	}
	expected := []string{
		"registry/deps-noscience:v2",
		"registry/deps-science:v2",
		"registry/deps-all:v2",
	}
	require.Equal(t, expected, config.BuildBaseImages())
}
