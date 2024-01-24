package internal

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Image struct {
	Registry string
	Name     string
	Tag      string
}

func (i Image) String() string {
	return i.Url()
}

func (i Image) Url() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Name, i.Tag)
}

func (i Image) Desc() (v1.Image, name.Reference, error) {
	return DescImage(i.Url())
}

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

func GetImageEnVarPrefix(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("unable to parse image name %s: %v", image, err)
	}
	repStr := ref.Context().RepositoryStr()
	var replacer = strings.NewReplacer("/", "_", "-", "_")
	prefix := strings.ToUpper(replacer.Replace(repStr))
	return prefix, nil
}
