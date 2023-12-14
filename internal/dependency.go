package internal

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
