package internal

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// IsBuiltFrom returns true if img was built on top of base. An image does
// not record the name of its base image, but "FROM base" makes it start with
// all the layers of base, and with its whole build history: the history also
// lists the instructions which create no layer (ENV, USER, ENTRYPOINT...),
// so it tells apart two bases with the same layers but another config.
// Only the manifests and configs are read, no layer is downloaded.
func IsBuiltFrom(img v1.Image, base v1.Image) (bool, error) {
	sameLayers, err := startsWithLayersOf(img, base)
	if err != nil || !sameLayers {
		return false, err
	}
	return startsWithHistoryOf(img, base)
}

func startsWithHistoryOf(img v1.Image, base v1.Image) (bool, error) {
	imgConfig, err := img.ConfigFile()
	if err != nil {
		return false, fmt.Errorf("unable to read image config: %v", err)
	}
	baseConfig, err := base.ConfigFile()
	if err != nil {
		return false, fmt.Errorf("unable to read base image config: %v", err)
	}
	if len(baseConfig.History) > len(imgConfig.History) {
		return false, nil
	}
	for i, baseEntry := range baseConfig.History {
		imgEntry := imgConfig.History[i]
		if !baseEntry.Created.Time.Equal(imgEntry.Created.Time) ||
			baseEntry.CreatedBy != imgEntry.CreatedBy ||
			baseEntry.Comment != imgEntry.Comment ||
			baseEntry.EmptyLayer != imgEntry.EmptyLayer {
			return false, nil
		}
	}
	return true, nil
}

func startsWithLayersOf(img v1.Image, base v1.Image) (bool, error) {
	imgLayers, err := img.Layers()
	if err != nil {
		return false, fmt.Errorf("unable to list image layers: %v", err)
	}
	baseLayers, err := base.Layers()
	if err != nil {
		return false, fmt.Errorf("unable to list base image layers: %v", err)
	}
	if len(baseLayers) > len(imgLayers) {
		return false, nil
	}
	for i, baseLayer := range baseLayers {
		baseDigest, err := baseLayer.Digest()
		if err != nil {
			return false, fmt.Errorf("unable to get base image layer digest: %v", err)
		}
		imgDigest, err := imgLayers[i].Digest()
		if err != nil {
			return false, fmt.Errorf("unable to get image layer digest: %v", err)
		}
		if baseDigest != imgDigest {
			return false, nil
		}
	}
	return true, nil
}

// IsBuiltFromUrl returns true if the image at imageUrl was built on top of
// the image at baseUrl, or if baseUrl is empty. If an image cannot be read,
// the image is reported as not built from the base, with an error, so that
// the caller rebuilds it rather than using a stale image.
func IsBuiltFromUrl(imageUrl string, baseUrl string) (bool, error) {
	if baseUrl == "" {
		return true, nil
	}
	img, _, err := DescImage(imageUrl)
	if err != nil {
		return false, err
	}
	base, _, err := DescImage(baseUrl)
	if err != nil {
		return false, err
	}
	return IsBuiltFrom(img, base)
}

// BaseImage returns the base image of a build of the project image: the
// image of the dependency with "baseOf: <build>". An empty build is the
// default build. If no dependency declares a "baseOf", it returns an empty
// string: the base image is not checked.
func (c ProjConfig) BaseImage(build string) (string, error) {
	declared := false
	base := ""
	if build == "" {
		build = c.DefaultBuild
	}
	for _, dep := range c.Dependencies {
		if dep.BaseOf == "" {
			continue
		}
		declared = true
		if dep.BaseOf != build {
			continue
		}
		if base != "" {
			return "", fmt.Errorf("build %q has two base images: %s and %s", build, base, dep.Image)
		}
		base = dep.Image
	}
	if !declared {
		return "", nil
	}
	if build == "" {
		return "", fmt.Errorf("no build given and no defaultBuild in configuration, unable to select a base image")
	}
	if base == "" {
		return "", fmt.Errorf("no dependency with \"baseOf: %s\" in configuration", build)
	}
	return base, nil
}

// ImageSuffix returns the suffix of the project image name for a build:
// the build name, except for the default build, whose image has no suffix
func (c ProjConfig) ImageSuffix(build string) string {
	if build == c.DefaultBuild {
		return ""
	}
	return build
}
