package internal

import (
	"fmt"

	"github.com/k8s-school/ciux/log"
)

type Project struct {
	Git    Git
	Config Config
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
func (project Project) GetDepsWorkBranch() (gitDeps []Git, err error) {
	gitMain := project.Git
	revMain, err := gitMain.GetRevision()
	if err != nil {
		return nil, fmt.Errorf("unable to describe git repository: %v", err)
	}
	gitDeps = []Git{}
	for _, dep := range project.Config.Dependencies {
		gitDep := Git{IsRemote: true, Url: dep.Url}
		hasBranch, err := gitDep.HasBranch(revMain.Branch)
		if err != nil {
			return nil, fmt.Errorf("unable to check branch existence for dependency repository %s: %v", gitDep.Url, err)
		}
		if hasBranch {
			gitDep.WorkBranch = revMain.Branch
		} else {
			// TODO Retrieve the default branch in GitLsRemote()
			main, err := gitMain.MainBranch()
			if err != nil {
				return nil, fmt.Errorf("unable to get main branch for project repository %s: %v", gitMain.Url, err)
			}
			gitDep.WorkBranch = main
		}
		log.Debugf("Repository: %s, work branch: %+v", gitDep.Url, gitDep.WorkBranch)
		gitDeps = append(gitDeps, gitDep)
	}
	return gitDeps, nil
}
