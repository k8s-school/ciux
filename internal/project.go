package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/k8s-school/ciux/log"
)

type Project struct {
	Git     Git
	Config  Config
	GitDeps []Git
}

// NewProject creates a new Project struct
// It reads the repository_path/.ciux.yaml configuration file
// and retrieve the work branch for all dependencies
func NewProject(repository_path string) Project {
	git, err := NewGit(repository_path)
	FailOnError(err)
	config, err := NewConfig(repository_path)
	FailOnError(err)
	p := Project{
		Git:    git,
		Config: config,
	}
	p.ScanRemoteDeps()
	return p
}

func (p *Project) String() string {

	name, err := p.Git.GetName()
	if err != nil {
		return fmt.Sprintf("unable to get project name: %v", err)
	}
	revMain, err := p.Git.GetRevision()
	if err != nil {
		return fmt.Sprintf("unable to describe project repository: %v", err)
	}
	rootMain, err := p.Git.GetRoot()
	if err != nil {
		return fmt.Sprintf("unable to get root of project repository: %v", err)
	}
	msg := fmt.Sprintf("Project %s\n  %s %+s\n", name, rootMain, revMain.GetVersion())
	msg += "Dependencies:"
	for _, dep := range p.GitDeps {
		revDep, err := dep.GetRevision()
		if err != nil {
			return msg + fmt.Sprintf("unable to describe git repository: %v", err)
		}
		rootDep, err := dep.GetRoot()
		if err != nil {
			return msg + fmt.Sprintf("unable to get root of git repository: %v", err)
		}
		msg += fmt.Sprintf("\n  %s %s", rootDep, revDep.GetVersion())
	}
	return msg
}

func (p *Project) SetDepsRepos(basePath string) error {
	for i, depConfig := range p.Config.Dependencies {
		if depConfig.Clone {
			singleBranch := true
			err := p.GitDeps[i].Clone(basePath, singleBranch)
			if err != nil {
				return fmt.Errorf("unable to set git repository %s: %v", p.GitDeps[i].Url, err)
			}
		}
	}
	return nil
}

func (p *Project) CheckImages() ([]name.Reference, error) {
	foundImages := []name.Reference{}
	for _, gitDep := range p.GitDeps {
		rev, err := gitDep.GetRevision()
		if err != nil {
			return foundImages, fmt.Errorf("unable to describe git repository: %v", err)
		}
		log.Debugf("Dep repo: %s, version: %+v", gitDep.Url, rev.GetVersion())
		// TODO: Set image path at configuration time
		depName, err := LastDir(gitDep.Url)
		if err != nil {
			return foundImages, fmt.Errorf("unable to get last directory of git repository: %v", err)
		}
		imageUrl := fmt.Sprintf("%s/%s:%s", p.Config.Registry, depName, rev.GetVersion())
		_, ref, err := DescImage(imageUrl)
		if err != nil {
			return foundImages, fmt.Errorf("unable to check image existence: %v, %v", err, ref)
		}
		foundImages = append(foundImages, ref)
	}
	return foundImages, nil
}

// ScanRemoteDeps retrieves the work branch for each dependency
// It is the same branch as the main repository if it exists
// or the default branch of the dependency repository otherwise
func (project *Project) ScanRemoteDeps() (err error) {
	gitMain := project.Git
	revMain, err := gitMain.GetRevision()
	if err != nil {
		return fmt.Errorf("unable to describe git repository: %v", err)
	}
	gitDeps := []Git{}
	for _, dep := range project.Config.Dependencies {
		gitDep := Git{Url: dep.Url}
		hasBranch, err := gitDep.HasBranch(revMain.Branch)
		if err != nil {
			return fmt.Errorf("unable to check branch existence for dependency repository %s: %v", gitDep.Url, err)
		}
		if hasBranch {
			gitDep.WorkBranch = revMain.Branch
		} else {
			// TODO Retrieve the default branch in GitLsRemote()
			main, err := gitMain.MainBranch()
			if err != nil {
				return fmt.Errorf("unable to get main branch for project repository %s: %v", gitMain.Url, err)
			}
			gitDep.WorkBranch = main
		}
		gitDeps = append(gitDeps, gitDep)

	}
	log.Debugf("gitDeps: %+v", gitDeps)
	project.GitDeps = gitDeps
	return nil
}

// WriteOutConfig writes out the shell configuration file
// used be the CI/CD pipeline
func (p *Project) WriteOutConfig() error {
	root, err := p.Git.GetRoot()
	if err != nil {
		return fmt.Errorf("unable to get root of git repository: %v", err)
	}

	ciuxConfig := filepath.Join(root, "ciux.sh")
	f, err := os.Create(ciuxConfig)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %v", ciuxConfig, err)
	}
	defer f.Close()

	for _, dep := range p.GitDeps {
		if !dep.isRemote() {
			varName, err := dep.GetEnVarName()
			if err != nil {
				return fmt.Errorf("unable to get environment variable name for git repository %v: %v", dep, err)
			}

			root, err := dep.GetRoot()
			if err != nil {
				return fmt.Errorf("unable to get root of git repository: %v", err)
			}

			depEnv := fmt.Sprintf("export %s=%s\n", varName, root)
			_, err = f.WriteString(depEnv)
			if err != nil {
				return fmt.Errorf("unable to write variable %s to file %s: %v", varName, ciuxConfig, err)
			}
		}
	}
	return nil
}
