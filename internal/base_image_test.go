package internal

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"
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

// withConfigOnlyChange returns base plus an ENV instruction: same layers,
// as an ENV, USER or ENTRYPOINT instruction creates no layer, but a new
// entry in the build history
func withConfigOnlyChange(t *testing.T, base v1.Image) v1.Image {
	cf, err := base.ConfigFile()
	require.NoError(t, err)
	cf = cf.DeepCopy()
	cf.Config.Env = append(cf.Config.Env, "SPARK_HOME=/opt/spark-3.5")
	cf.History = append(cf.History, v1.History{
		Created:    v1.Time{Time: time.Now()},
		CreatedBy:  "ENV SPARK_HOME=/opt/spark-3.5",
		EmptyLayer: true,
	})
	img, err := mutate.ConfigFile(base, cf)
	require.NoError(t, err)
	return img
}

func TestIsBuiltFrom(t *testing.T) {
	base := randomImage(t)
	img := builtOn(t, base)
	newBase := withConfigOnlyChange(t, base)

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
		{"base changed only its config", img, newBase, false},
		{"built on a base with a config-only change", builtOn(t, newBase), newBase, true},
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

func TestIsBuiltFromUrl(t *testing.T) {
	server := httptest.NewServer(registry.New())
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")

	// As in fink-broker, the science base is built on the noscience one
	noscienceBase := randomImage(t)
	scienceBase := builtOn(t, noscienceBase)
	urlNoscience := pushImage(t, host, "deps-noscience", noscienceBase)
	urlScience := pushImage(t, host, "deps-science", scienceBase)
	urlImage := pushImage(t, host, "broker", builtOn(t, scienceBase))
	urlMissing := fmt.Sprintf("%s/deps-missing:v1", host)

	tests := []struct {
		name      string
		base      string
		expected  bool
		expectErr bool
	}{
		{"built on the base", urlScience, true, false},
		{"built on another base", pushImage(t, host, "other", randomImage(t)), false, false},
		{"no base declared", "", true, false},
		{"base missing from the registry", urlMissing, false, true},
		// The noscience base is a prefix of the science one: the image
		// is built on it too, the build must tell which base to check
		{"built on a base of the base", urlNoscience, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBuiltFromUrl(urlImage, tt.base)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestBaseImage(t *testing.T) {
	deps := []DepConfig{
		{Image: "registry/deps-noscience:v2", BaseOf: "noscience"},
		{Image: "registry/deps-science:v2", BaseOf: "science"},
		{Image: "registry/hadoop:v1"},
		{Url: "https://github.com/org/repo"},
	}
	tests := []struct {
		name      string
		config    ProjConfig
		build     string
		expected  string
		expectErr bool
	}{
		{"base of the build", ProjConfig{Dependencies: deps}, "noscience", "registry/deps-noscience:v2", false},
		{"no build: base of the default build", ProjConfig{Dependencies: deps, DefaultBuild: "science"}, "", "registry/deps-science:v2", false},
		{"no build and no default build", ProjConfig{Dependencies: deps}, "", "", true},
		{"unknown build", ProjConfig{Dependencies: deps}, "sciense", "", true},
		{"no base declared: no base check", ProjConfig{Dependencies: deps[2:]}, "science", "", false},
		{"no base declared and no build", ProjConfig{Dependencies: deps[2:]}, "", "", false},
		{"two bases for one build", ProjConfig{Dependencies: append([]DepConfig{
			{Image: "registry/deps-science:v1", BaseOf: "science"}}, deps...)}, "science", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.BaseImage(tt.build)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestImageSuffix(t *testing.T) {
	config := ProjConfig{DefaultBuild: "science"}
	require.Equal(t, "", config.ImageSuffix("science"))
	require.Equal(t, "", config.ImageSuffix(""))
	require.Equal(t, "noscience", config.ImageSuffix("noscience"))
	require.Equal(t, "science", ProjConfig{}.ImageSuffix("science"))
}
