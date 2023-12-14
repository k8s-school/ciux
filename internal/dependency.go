package internal

import "fmt"

type Dependency struct {
	Clone   bool
	Git     *Git
	Image   string
	Pull    bool
	Package string
}

func (dep *Dependency) String() string {
	if dep.Package != "" {
		return dep.Package
	} else if dep.Image != "" {
		return dep.Image
	} else {
		return dep.Git.Url
	}
}

func (dep *Dependency) GetImageName(imageRegistry string) (string, error) {
	if dep.Image != "" {
		return dep.Image, nil
	} else {
		gitDep := dep.Git
		rev, err := gitDep.GetRevision()
		if err != nil {
			return "", fmt.Errorf("unable to describe git repository: %v", err)
		}
		// TODO: Set image path at configuration time
		depName, err := LastDir(gitDep.Url)
		if err != nil {
			return "", fmt.Errorf("unable to get last directory of git repository: %v", err)
		}
		imageUrl := fmt.Sprintf("%s/%s:%s", imageRegistry, depName, rev.GetVersion())
		return imageUrl, nil
	}
}
