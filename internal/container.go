package internal

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func ListTags(src string) ([]string, error) {
	repo, err := name.NewRepository(src)
	if err != nil {
		return nil, fmt.Errorf("parsing repo %q: %w", src, err)
	}
	return remote.List(repo)
}

func DescImage(r string) (v1.Image, name.Reference, error) {
	ref, err := name.ParseReference(r)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reference %q: %w", r, err)
	}
	img, err := remote.Image(ref)
	if err != nil {
		return nil, nil, fmt.Errorf("reading image %q: %w", ref, err)
	}
	return img, ref, nil
}
