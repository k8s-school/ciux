package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/k8s-school/ciux/log"
	"k8s.io/apimachinery/pkg/labels"
)

type Project struct {
	GitMain       *Git
	SourcePathes  []string
	ImageRegistry string
	Image         Image
	Dependencies  []*Dependency
	// Required for github actions, which fetch a single commit by default
	ForcedBranch      string
	TemporaryRegistry string
}

func NewCoreProject(repository_path string, forcedBranch string) (Project, ProjConfig, error) {
	git, err := NewGit(repository_path)
	if err != nil {
		return Project{}, ProjConfig{}, fmt.Errorf("unable to create git repository: %v", err)
	}
	config, err := NewConfig(repository_path)
	if err != nil {
		return Project{}, ProjConfig{}, fmt.Errorf("unable to read configuration file: %v", err)
	}
	p := Project{
		GitMain:       git,
		SourcePathes:  config.SourcePathes,
		ImageRegistry: config.Registry,
		ForcedBranch:  forcedBranch,
	}
	return p, config, nil
}

// NewProject creates a new Project struct
// It reads the repository_path/.ciux.yaml configuration file
// and retrieve the work branch for all dependencies
func NewProject(repository_path string, forcedBranch string, labelSelector string) (Project, error) {

	p, config, err := NewCoreProject(repository_path, forcedBranch)
	if err != nil {
		return Project{}, err
	}

	selectors := labels.Everything()
	if len(labelSelector) > 0 {
		selectors, err = labels.Parse(labelSelector)
		if err != nil {
			return Project{}, fmt.Errorf("unable to parse label selector: %v", err)
		}
	}

	deps := []*Dependency{}
	for _, depConfig := range config.Dependencies {
		var dep *Dependency
		if depConfig.Package != "" {
			dep = &Dependency{
				Package: depConfig.Package,
			}
		} else if depConfig.Image != "" {
			dep = &Dependency{
				Image: depConfig.Image,
			}
		} else {
			dep = &Dependency{
				Clone: depConfig.Clone,
				Pull:  depConfig.Pull,
				Git: &Git{
					Url: depConfig.Url,
				},
			}
		}
		if selectors.Matches(depConfig.Labels) {
			slog.Debug("Dependency selected", "labels", depConfig.Labels, "dep", dep)
			deps = append(deps, dep)
		}
	}

	p.Dependencies = deps
	p.scanRemoteDeps()
	return p, nil
}

func (p *Project) String() string {

	name, err := p.GitMain.GetName()
	if err != nil {
		return fmt.Sprintf("unable to get project name: %v", err)
	}
	revMain, err := p.GitMain.GetHeadRevision()
	if err != nil {
		return fmt.Sprintf("unable to describe project repository: %v", err)
	}
	rootMain, err := p.GitMain.GetRoot()
	if err != nil {
		return fmt.Sprintf("unable to get root of project repository: %v", err)
	}
	msg := fmt.Sprintf("Project %s\n  %s@%s\n", name, rootMain, revMain.GetVersion())
	if len(p.Dependencies) != 0 {
		msg += "Dependencies:"
		for _, dep := range p.Dependencies {

			if dep.Package != "" {
				msg += fmt.Sprintf("\n  Package: %s", dep.Package)
			} else if dep.Image != "" {
				msg += fmt.Sprintf("\n  Image: %s", dep.Image)
			} else if dep.Git != nil {
				slog.Debug("Dependency", "url", dep.Git.Url, "branch", dep.Git.WorkBranch)
				if !dep.Git.isRemoteOnly() {
					revDep, err := dep.Git.GetHeadRevision()
					if err != nil {
						return msg + fmt.Sprintf("unable to describe git repository: %v", err)
					}
					rootDep, err := dep.Git.GetRoot()
					if err != nil {
						return msg + fmt.Sprintf("unable to get root of git repository: %v", err)
					}
					// in-place means that the dependency was in-place before project ignition
					msg += fmt.Sprintf("\n  %s %s in-place=%t", rootDep, revDep.GetVersion(), dep.Git.InPlace)
				} else {
					msg += fmt.Sprintf("\n  %s remote-only=true branch=%s", dep.Git.Url, dep.Git.WorkBranch)
				}
				if dep.Pull {
					msg += " pull=true"
				}
			}
		}
	}
	return msg
}

func (p *Project) RetrieveDepsSources(basePath string) error {
	for i, dep := range p.Dependencies {
		if dep.Clone {
			singleBranch := true
			err := p.Dependencies[i].Git.CloneOrOpen(basePath, singleBranch)
			if err != nil {
				return fmt.Errorf("unable to set git repository %s: %v", p.Dependencies[i].Git.Url, err)
			}
		}
	}
	return nil
}

func (p *Project) AddInPlaceDepsSources(basePath string) error {
	for i, dep := range p.Dependencies {
		if dep.Clone {
			err := p.Dependencies[i].Git.OpenIfExists(basePath)
			if err != nil {
				return fmt.Errorf("unable to set git repository %s: %v", p.Dependencies[i].Git.Url, err)
			}
		}
	}
	return nil
}

func (p *Project) CheckDepImages() ([]name.Reference, error) {
	foundImages := []name.Reference{}
	for i, dep := range p.Dependencies {
		if dep.Pull {
			imageUrl, err := dep.GetImageName(p.ImageRegistry)
			if err != nil {
				return foundImages, fmt.Errorf("unable to get image name for git repository %s: %v", p.Dependencies[i].Git.Url, err)
			}
			_, ref, err := DescImage(imageUrl)
			if err != nil {
				return foundImages, fmt.Errorf("unable to check image existence: %v, %v", err, ref)
			}
			foundImages = append(foundImages, ref)
		} else if dep.Image != "" {
			_, ref, err := DescImage(dep.Image)
			if err != nil {
				return foundImages, fmt.Errorf("unable to check image existence: %v, %v", err, ref)
			}
			foundImages = append(foundImages, ref)
		}
	}
	return foundImages, nil
}

func (p *Project) InstallGoModules() (string, error) {
	msg := ""
	for _, dep := range p.Dependencies {
		if dep.Package != "" {
			cmd := fmt.Sprintf("go install %s", dep.Package)
			outstr, errstr, err := ExecCmd(cmd, false, false)
			slog.Debug("Install package", "cmd", cmd, "out", outstr, "err", errstr)
			if err != nil {
				return msg, fmt.Errorf("unable to install go module %s: %v", dep.Package, err)
			}
			msg += fmt.Sprintf("  %s\n", dep.Package)
		} else if dep.Clone {
			isGoMod, err := dep.Git.IsGoModule()
			if err != nil {
				return msg, fmt.Errorf("unable to check if git repository %s is a go module: %v", dep.Git.Url, err)
			}
			if isGoMod {
				err := dep.Git.GoInstall()
				if err != nil {
					return msg, fmt.Errorf("unable to install go modules for git repository %s: %v", dep.Git.Url, err)
				}
				msg += fmt.Sprintf("  %s from-src=true\n", dep.Git.Url)
			}
		}
	}
	return msg, nil
}

// scanRemoteDeps retrieves the work branch for each dependency
// It is the same branch as the main repository if it exists
// or the default branch of the dependency repository otherwise
func (project *Project) scanRemoteDeps() error {

	var err error

	if project.ForcedBranch == "" {
		project.GitMain.WorkBranch, err = project.GitMain.GetBranch()
		if err != nil {
			return fmt.Errorf("unable to get work branch for project main repository: %v", err)
		}
	} else {
		project.GitMain.WorkBranch = project.ForcedBranch
	}

	var hash string
	for _, dep := range project.Dependencies {
		if dep.Git != nil {
			err = dep.Git.LsRemote()
			if err != nil {
				return fmt.Errorf("unable to ls-remote for dependency repository %s: %v", dep.Git.Url, err)
			}
			var hasBranch bool
			hasBranch, hash, err = dep.Git.HasBranch(project.GitMain.WorkBranch)
			if err != nil {
				return fmt.Errorf("unable to check branch existence for dependency repository %s: %v", dep.Git.Url, err)
			}
			if hasBranch {
				dep.Git.WorkBranch = project.GitMain.WorkBranch
			} else {
				var main string
				main, hash, err = dep.Git.MainBranch()
				if err != nil {
					return fmt.Errorf("unable to get main branch for project repository %s: %v", project.GitMain.Url, err)
				}
				dep.Git.WorkBranch = main
			}
			dep.Git.RemoteHash = hash
		}
	}
	if log.IsDebugEnabled() {
		for _, dep := range project.Dependencies {
			if dep.Git != nil {
				slog.Debug("Dependency", "url", dep.Git.Url, "branch", dep.Git.WorkBranch)
			}
		}
	}
	return nil
}

// WriteOutConfig writes out the shell configuration file
// used be the CI/CD pipeline
func (p *Project) WriteOutConfig() (string, error) {
	msg := ""
	var ciuxConfigFile = os.Getenv("CIUXCONFIG")
	if len(ciuxConfigFile) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return msg, fmt.Errorf("unable to get user home directory: %v", err)
		}
		ciuxCfgDir := filepath.Join(home, ".ciux")
		os.MkdirAll(ciuxCfgDir, 0755)
		if err != nil {
			return msg, fmt.Errorf("unable to create directory %s: %v", ciuxCfgDir, err)
		}
		ciuxConfigFile = filepath.Join(ciuxCfgDir, "ciux.sh")
	}
	f, err := os.Create(ciuxConfigFile)
	if err != nil {
		return msg, fmt.Errorf("unable to create configuration file %s: %v", ciuxConfigFile, err)
	}
	defer f.Close()

	gitDeps := []*Git{}
	imageDeps := []string{}
	for _, dep := range p.Dependencies {
		if dep.Git != nil {
			gitDeps = append(gitDeps, dep.Git)
		} else if dep.Image != "" {
			imageDeps = append(imageDeps, dep.Image)
		}
	}

	gitRepos := append(gitDeps, p.GitMain)
	for _, gitObj := range gitRepos {
		varName, err := gitObj.GetEnVarPrefix()
		if err != nil {
			return msg, fmt.Errorf("unable to get environment variable name for git repository %v: %v", gitObj, err)
		}
		if !gitObj.isRemoteOnly() {
			root, err := gitObj.GetRoot()
			if err != nil {
				return msg, fmt.Errorf("unable to get root of git repository: %v", err)
			}

			depEnv := fmt.Sprintf("export %s_DIR=%s\n", varName, root)
			_, err = f.WriteString(depEnv)
			if err != nil {
				return msg, fmt.Errorf("unable to write variable %s to file %s: %v", varName, ciuxConfigFile, err)
			}

			rev, err := gitObj.GetHeadRevision()
			if err != nil {
				return msg, fmt.Errorf("unable to describe git repository: %v", err)
			}
			depVersion := fmt.Sprintf("export %s_VERSION=%s\n", varName, rev.GetVersion())
			_, err = f.WriteString(depVersion)
			if err != nil {
				return msg, fmt.Errorf("unable to write variable %s_VERSION to file %s: %v", varName, ciuxConfigFile, err)
			}

		}

		depEnv := fmt.Sprintf("export %s_WORKBRANCH=%s\n", varName, gitObj.WorkBranch)
		_, err = f.WriteString(depEnv)
		if err != nil {
			return msg, fmt.Errorf("unable to write variable %s to file %s: %v", varName, ciuxConfigFile, err)
		}
	}

	for _, image := range imageDeps {
		varName, err := GetImageEnVarPrefix(image)
		if err != nil {
			return msg, fmt.Errorf("unable to get environment variable name for image %s: %v", image, err)
		}
		imageEnv := fmt.Sprintf("export %s_IMAGE=%s\n", varName, image)
		_, err = f.WriteString(imageEnv)
		if err != nil {
			return msg, fmt.Errorf("unable to write variable %s_IMAGE to file %s: %v", varName, ciuxConfigFile, err)
		}
	}
	imageRegistry := p.ImageRegistry
	imageEnv := fmt.Sprintf("export CIUX_IMAGE_REGISTRY=%s\n", imageRegistry)
	_, err = f.WriteString(imageEnv)
	if err != nil {
		return msg, fmt.Errorf("unable to write variable CIUX_IMAGE_REGISTRY to file %s: %v", ciuxConfigFile, err)
	}
	imageName := p.Image.Name
	imageEnv = fmt.Sprintf("export CIUX_IMAGE_NAME=%s\n", imageName)
	_, err = f.WriteString(imageEnv)
	if err != nil {
		return msg, fmt.Errorf("unable to write variable CIUX_IMAGE_NAME to file %s: %v", ciuxConfigFile, err)
	}
	imageTag := p.Image.Tag
	prefix, err := p.GitMain.GetEnVarPrefix()
	if err != nil {
		return msg, fmt.Errorf("unable to get environment variable prefix for project main git repository: %v", err)
	}
	imageEnv = fmt.Sprintf("# Similar to %s_VERSION\n", prefix)
	imageEnv += fmt.Sprintf("export CIUX_IMAGE_TAG=%s\n", imageTag)
	_, err = f.WriteString(imageEnv)
	if err != nil {
		return msg, fmt.Errorf("unable to write variable CIUX_IMAGE_TAG to file %s: %v", ciuxConfigFile, err)
	}

	imageUrl := fmt.Sprintf("export CIUX_IMAGE_URL=%s\n", p.Image.Url())
	_, err = f.WriteString(imageUrl)
	if err != nil {
		return msg, fmt.Errorf("unable to write variable CIUX_IMAGE_URL to file %s: %v", ciuxConfigFile, err)
	}
	notInRegistry := fmt.Sprintf("export CIUX_BUILD=%t\n", !p.Image.InRegistry)
	_, err = f.WriteString(notInRegistry)
	if err != nil {
		return msg, fmt.Errorf("unable to write variable CIUX_BUILD to file %s: %v", ciuxConfigFile, err)
	}

	msg = fmt.Sprintf("Configuration file:\n  %s", ciuxConfigFile)
	return msg, nil
}

func (p *Project) GetGits() []*Git {
	gits := []*Git{p.GitMain}
	for _, dep := range p.Dependencies {
		if dep.Git != nil {
			gits = append(gits, dep.Git)
		}
	}
	return gits
}

// GetImageName compute the name and tag for the project image
//
//	it checks the git repository for changes
//	a suffix can be added to image name
//	image existence in the registry can be checked
func (project *Project) GetImageName(suffix string, checkRegistry bool) error {
	gitMain := project.GitMain

	slog.Debug("Project source directories", "sourcePathes", project.SourcePathes)

	head, err := gitMain.Repository.Head()
	if err != nil {
		return fmt.Errorf("unable to get HEAD of repository %s: %v", gitMain.Url, err)
	}
	hashes, err := FindCodeChange(gitMain.Repository, head.Hash(), project.SourcePathes)
	if err != nil {
		return fmt.Errorf("unable to find code change in repository %s: %v", gitMain.Url, err)
	}
	if len(hashes) != 0 {
		rev, err := gitMain.GetRevision(hashes[0])
		if err != nil {
			return fmt.Errorf("unable to describe git repository: %v", err)
		}
		slog.Info("Project image with latest code changes", "hash", hashes[0], "version", rev.GetVersion())
	}
	imageName, err := gitMain.GetName()
	if err != nil {
		return fmt.Errorf("unable to get project name: %v", err)
	}
	if len(suffix) > 0 {
		imageName = fmt.Sprintf("%s-%s", imageName, suffix)
	}

	image := Image{
		Registry: project.ImageRegistry,
		Name:     imageName,
	}
	if checkRegistry {
		inRegistryImage, err := project.findInRegistryImage(imageName, hashes)
		// NOTE: it is difficult to filter the error here (network, image not exist, etc...)
		// so in case of error we consider the image not in the registry
		if inRegistryImage == nil || err != nil {
			image.InRegistry = false
			slog.Debug("Image not found in registry", "image", image)
		} else {
			image = *inRegistryImage
		}
	} else {
		image.InRegistry = false
	}

	if !image.InRegistry {
		rev, err1 := gitMain.GetHeadRevision()
		if err1 != nil {
			return fmt.Errorf("unable to describe git repository: %v", err1)
		}
		image.Tag = rev.GetVersion()
		if project.TemporaryRegistry != "" {
			image.Registry = project.TemporaryRegistry
		}
	}

	project.Image = image
	return nil
}

// findInRegistryImage returns the first image which exist in the registry for a commit hash which is in hashes[]
func (project *Project) findInRegistryImage(imageName string, hashes []plumbing.Hash) (*Image, error) {
	gitMain := project.GitMain
	image := Image{
		Registry: project.ImageRegistry,
		Name:     imageName,
	}

	for _, hash := range hashes {
		rev, err := gitMain.GetRevision(hash)
		if err != nil {
			return nil, fmt.Errorf("unable to describe git repository for commit %v: %v", hash, err)
		}
		image.Tag = rev.GetVersion()
		slog.Debug("Check image in registry", "image", image)
		_, _, errRegistry := image.Desc()
		if errRegistry != nil {
			image.InRegistry = false
		} else {
			image.InRegistry = true
			slog.Debug("Found image in registry", "image", image)
			break
		}
	}
	if image.InRegistry {
		return &image, nil
	} else {
		return nil, nil
	}
}
