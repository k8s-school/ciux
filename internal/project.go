package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k8s-school/ciux/log"
)

type Project struct {
	Git     Git
	Config  Config
	GitDeps []Git
}

func NewProject(repository_path string) Project {
	git, err := NewGit(repository_path)
	FailOnError(err)
	config, err := NewConfig(repository_path)
	FailOnError(err)
	return Project{
		Git:    git,
		Config: config,
	}
}

// GetDepsWorkBranch returns a list of Git structures which contain
// the work branch for each dependency in Git.Revision.Branch
// It is the same branch as the main repository if it exists
// or the default branch of the dependency repository otherwise
func (project Project) GetDepsWorkBranch() (err error) {
	gitMain := project.Git
	revMain, err := gitMain.GetRevision()
	if err != nil {
		return fmt.Errorf("unable to describe git repository: %v", err)
	}
	project.GitDeps = []Git{}
	for _, dep := range project.Config.Dependencies {
		gitDep := Git{IsRemote: true, Url: dep.Url}
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
		log.Debugf("Repository: %s, work branch: %+v", gitDep.Url, gitDep.WorkBranch)
		project.GitDeps = append(project.GitDeps, gitDep)
	}
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

		name, err := dep.GetName()
		if err != nil {
			return fmt.Errorf("unable to get name from url %s: %v", dep.Url, err)
		}

		root, err := dep.GetRoot()
		if err != nil {
			return fmt.Errorf("unable to get root of git repository: %v", err)
		}

		name = strings.ToUpper(name)
		depEnv := fmt.Sprintf("export %s=%s\n", name, root)
		_, err = f.WriteString(depEnv)
		if err != nil {
			return fmt.Errorf("unable to write variable %s to file %s: %v", name, ciuxConfig, err)
		}
	}
	return nil
}
