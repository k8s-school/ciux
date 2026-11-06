package internal

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// IsBuiltFrom returns true if img was built on top of base, i.e. if the
// layers of base are the first layers of img. An image does not record the
// name of its base image, but "FROM base" makes it start with all its layers.
// Only the manifests are read, no layer is downloaded.
func IsBuiltFrom(img v1.Image, base v1.Image) (bool, error) {
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

// IsBuiltFromAny returns true if the image at imageUrl was built on top of
// one of the images at baseUrls, or if baseUrls is empty. An image which
// cannot be checked is reported as not built from them, with an error, so
// that the caller rebuilds it rather than using a stale image.
func IsBuiltFromAny(imageUrl string, baseUrls []string) (bool, error) {
	if len(baseUrls) == 0 {
		return true, nil
	}
	img, _, err := DescImage(imageUrl)
	if err != nil {
		return false, err
	}
	for _, baseUrl := range baseUrls {
		base, _, err := DescImage(baseUrl)
		if err != nil {
			return false, err
		}
		built, err := IsBuiltFrom(img, base)
		if err != nil {
			return false, err
		}
		if built {
			return true, nil
		}
	}
	return false, nil
}

// BuildBaseImages returns the images of the dependencies labelled "build":
// the candidate base images of the project image
func (c ProjConfig) BuildBaseImages() []string {
	images := []string{}
	for _, dep := range c.Dependencies {
		if dep.Image != "" && dep.Labels.Has("build") {
			images = append(images, dep.Image)
		}
	}
	return images
}
