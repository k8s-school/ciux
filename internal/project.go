package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/k8s-school/ciux/log"
)

type Project struct {
	GitMain       Git
	ImageRegistry string
	Dependencies  []*Dependency
}

type Dependency struct {
	Clone bool
	Pull  bool
	Git   Git
}

// NewProject creates a new Project struct
// It reads the repository_path/.ciux.yaml configuration file
// and retrieve the work branch for all dependencies
func NewProject(repository_path string) Project {
	git, err := NewGit(repository_path)
	FailOnError(err)
	config, err := NewConfig(repository_path)
	FailOnError(err)

	deps := []*Dependency{}
	for _, depConfig := range config.Dependencies {
		dep := Dependency{
			Clone: depConfig.Clone,
			Pull:  depConfig.Pull,
			Git:   Git{Url: depConfig.Url},
		}
		deps = append(deps, &dep)
	}

	p := Project{
		GitMain:       git,
		ImageRegistry: config.Registry,
		Dependencies:  deps,
	}
	p.ScanRemoteDeps()
	return p
}

func (p *Project) String() string {

	name, err := p.GitMain.GetName()
	if err != nil {
		return fmt.Sprintf("unable to get project name: %v", err)
	}
	revMain, err := p.GitMain.GetRevision()
	if err != nil {
		return fmt.Sprintf("unable to describe project repository: %v", err)
	}
	rootMain, err := p.GitMain.GetRoot()
	if err != nil {
		return fmt.Sprintf("unable to get root of project repository: %v", err)
	}
	msg := fmt.Sprintf("Project %s\n  %s %+s\n", name, rootMain, revMain.GetVersion())
	msg += "Dependencies:"
	for _, dep := range p.Dependencies {
		revDep, err := dep.Git.GetRevision()
		if err != nil {
			return msg + fmt.Sprintf("unable to describe git repository: %v", err)
		}
		rootDep, err := dep.Git.GetRoot()
		if err != nil {
			return msg + fmt.Sprintf("unable to get root of git repository: %v", err)
		}
		msg += fmt.Sprintf("\n  %s %s", rootDep, revDep.GetVersion())
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

func (p *Project) CheckImages() ([]name.Reference, error) {
	foundImages := []name.Reference{}
	for i, dep := range p.Dependencies {
		if dep.Pull {
			gitDep := p.Dependencies[i].Git
			rev, err := gitDep.GetRevision()
			if err != nil {
				return foundImages, fmt.Errorf("unable to describe git repository: %v", err)
			}
			slog.Debug("Dep repo: %s, version: %+v", gitDep.Url, rev.GetVersion())
			// TODO: Set image path at configuration time
			depName, err := LastDir(gitDep.Url)
			if err != nil {
				return foundImages, fmt.Errorf("unable to get last directory of git repository: %v", err)
			}
			imageUrl := fmt.Sprintf("%s/%s:%s", p.ImageRegistry, depName, rev.GetVersion())
			_, ref, err := DescImage(imageUrl)
			if err != nil {
				return foundImages, fmt.Errorf("unable to check image existence: %v, %v", err, ref)
			}
			foundImages = append(foundImages, ref)
		}
	}
	return foundImages, nil
}

func (p *Project) InstallGoModules() error {
	for _, dep := range p.Dependencies {
		if dep.Clone {
			isGoMod, err := dep.Git.IsGoModule()
			if err != nil {
				return fmt.Errorf("unable to check if git repository %s is a go module: %v", dep.Git.Url, err)
			}
			if isGoMod {
				err := dep.Git.GoInstall()
				if err != nil {
					return fmt.Errorf("unable to install go modules for git repository %s: %v", dep.Git.Url, err)
				}
			}
		}
	}
	return nil
}

// ScanRemoteDeps retrieves the work branch for each dependency
// It is the same branch as the main repository if it exists
// or the default branch of the dependency repository otherwise
func (project *Project) ScanRemoteDeps() error {
	gitMain := project.GitMain
	revMain, err := gitMain.GetRevision()
	if err != nil {
		return fmt.Errorf("unable to describe git repository: %v", err)
	}
	for _, dep := range project.Dependencies {
		gitDep := &dep.Git
		hasBranch, err := gitDep.HasBranch(revMain.Branch)
		if err != nil {
			return fmt.Errorf("unable to check branch existence for dependency repository %s: %v", gitDep.Url, err)
		}
		if hasBranch {
			gitDep.WorkBranch = revMain.Branch
		} else {
			main, err := gitDep.MainBranch()
			if err != nil {
				return fmt.Errorf("unable to get main branch for project repository %s: %v", gitMain.Url, err)
			}
			gitDep.WorkBranch = main
		}
	}
	if log.IsDebugEnabled() {
		for _, dep := range project.Dependencies {
			slog.Debug("Dependency", "url", dep.Git.Url, "branch", dep.Git.WorkBranch)
		}
	}
	return nil
}

// WriteOutConfig writes out the shell configuration file
// used be the CI/CD pipeline
func (p *Project) WriteOutConfig() error {

	var ciuxConfigFile = os.Getenv("CIUXCONFIG")
	if len(ciuxConfigFile) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get user home directory: %v", err)
		}
		ciuxCfgDir := filepath.Join(home, ".ciux")
		os.MkdirAll(ciuxCfgDir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create directory %s: %v", ciuxCfgDir, err)
		}
		ciuxConfigFile = filepath.Join(ciuxCfgDir, "ciux.sh")
	}
	f, err := os.Create(ciuxConfigFile)
	if err != nil {
		return fmt.Errorf("unable to create configuration file %s: %v", ciuxConfigFile, err)
	}
	defer f.Close()

	gitDeps := []Git{}
	for _, dep := range p.Dependencies {
		gitDeps = append(gitDeps, dep.Git)
	}

	gitRepos := append(gitDeps, p.GitMain)
	for _, gitObj := range gitRepos {
		if !gitObj.isRemote() {
			varName, err := gitObj.GetEnVarPrefix()
			if err != nil {
				return fmt.Errorf("unable to get environment variable name for git repository %v: %v", gitObj, err)
			}

			root, err := gitObj.GetRoot()
			if err != nil {
				return fmt.Errorf("unable to get root of git repository: %v", err)
			}

			depEnv := fmt.Sprintf("export %s_DIR=%s\n", varName, root)
			_, err = f.WriteString(depEnv)
			if err != nil {
				return fmt.Errorf("unable to write variable %s to file %s: %v", varName, ciuxConfigFile, err)
			}

			rev, err := gitObj.GetRevision()
			if err != nil {
				return fmt.Errorf("unable to describe git repository: %v", err)
			}
			depVersion := fmt.Sprintf("export %s_VERSION=%s\n", varName, rev.GetVersion())
			_, err = f.WriteString(depVersion)
			if err != nil {
				return fmt.Errorf("unable to write variable %s_VERSION to file %s: %v", varName, ciuxConfigFile, err)
			}

		}
	}
	return nil
}

func (p *Project) GetGits() []Git {
	gits := []Git{p.GitMain}
	for _, dep := range p.Dependencies {
		gits = append(gits, dep.Git)
	}
	return gits
}
