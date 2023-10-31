package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/k8s-school/ciux/log"
)

type GitRevision struct {
	TagName  string
	Counter  int
	HeadHash string
	Dirty    bool
	Branch   string
}

type Git struct {
	Tags       []string
	Branches   []string
	Url        string
	Repository *git.Repository
	Revision   GitRevision
	Author     string
	IsRemote   bool
}

// GitSemverTagMap ...
func GitSemverTagMap(repo git.Repository) (*map[plumbing.Hash]*plumbing.Reference, error) {
	tagIter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	tagMap := map[plumbing.Hash]*plumbing.Reference{}
	err = tagIter.ForEach(func(r *plumbing.Reference) error {
		obj, err := repo.TagObject(r.Hash())
		switch err {
		case nil:
			if SemVerParse(r.Name().Short()) == nil {
				// Filter out tags that are not semver
				return nil
			}
			c, err := obj.Commit()
			if err != nil {
				return err
			}
			tagMap[c.Hash] = r
		case plumbing.ErrObjectNotFound:
			// Not an annotated tag object
			return nil
		default:
			// Some other error
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &tagMap, nil
}

// GitLsRemote returns branches and tag of a remote repository
// https://github.com/go-git/go-git/blob/master/_examples/ls-remote/main.go
func GitLsRemote(url string) (repo *Git, err error) {
	repo = &Git{IsRemote: true, Url: url}

	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo.Url},
	})

	refs, err := remote.List(&git.ListOptions{
		// Returns all references, including peeled references.
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list remote references: %v", err)
	}

	// Find annotated tags
	// the one with ^{} is the annotated tag
	// 2023/10/13 18:18:06 Tags found: v0.1
	// 2023/10/13 18:18:06 Tags found: lsst-france-meeting-may-2014
	// 2023/10/13 18:18:06 Tags found: test
	// 2023/10/13 18:18:06 Tags found: v0.5
	// 2023/10/13 18:18:06 Tags found: v0.1^{}
	// 2023/10/13 18:18:06 Tags found: v0.5^{}

	for _, ref := range refs {
		if ref.Name().IsBranch() {
			repo.Branches = append(repo.Branches, ref.Name().Short())
		}
		if ref.Name().IsTag() {
			repo.Tags = append(repo.Tags, ref.Name().Short())
		}
	}
	return repo, nil
}

func (repo *Git) HasBranch(branchname string) (bool, error) {
	found := false
	if repo.IsRemote {
		remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
			Name: "origin",
			URLs: []string{repo.Url},
		})
		refs, err := remote.List(&git.ListOptions{
			// Returns all references, including peeled references.
			PeelingOption: git.AppendPeeled,
		})
		if err != nil {
			return false, fmt.Errorf("unable to list remote references: %v", err)
		}
		for _, ref := range refs {
			if ref.Name().IsBranch() {
				if ref.Name().Short() == branchname {
					found = true
					break
				}
			}
		}
	} else {
		bIter, err := repo.Repository.Branches()
		if err != nil {
			return false, fmt.Errorf("unable to get branches: %v", err)
		}

		err = bIter.ForEach(func(c *plumbing.Reference) error {
			if c.Name().Short() == branchname {
				// Exit the loop
				found = true
				return storer.ErrStop
			}
			return nil
		})
		if err != nil {
			return false, fmt.Errorf("unable to loop on branches: %v", err)
		}
	}

	return found, nil
}

// Describe the reference as 'git describe ' will do
func (g *Git) Describe() error {

	head, err := g.Repository.Head()
	if err != nil {
		return fmt.Errorf("unable to find head: %v", err)
	}
	w, err := g.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("unable to find worktree: %v", err)
	}
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("unable to find worktree status: %v", err)
	}
	branchName := head.Name().Short()
	headHash := head.Hash().String()
	var dirty bool
	if status.IsClean() {
		dirty = false
	} else {
		dirty = true
	}

	// Fetch the reference log
	cIter, err := g.Repository.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return fmt.Errorf("unable to get reference log: %v", err)
	}

	// Build the semver tag map
	semverTags, err := GitSemverTagMap(*g.Repository)
	if err != nil {
		return fmt.Errorf("unable to get semver tags: %v", err)
	}

	// Search the latest semver tag
	var tag *plumbing.Reference
	var count int
	err = cIter.ForEach(func(c *object.Commit) error {
		ref, found := (*semverTags)[c.Hash]
		if found {
			tag = ref
			if tag != nil {
				// Exit the loop
				return storer.ErrStop
			} else {
				return fmt.Errorf("inconsistent semver tag map")
			}
		}
		count++
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to loop on commits: %v", err)
	}
	g.Revision = GitRevision{
		TagName:  tag.Name().Short(),
		Counter:  count,
		HeadHash: headHash,
		Dirty:    dirty,
		Branch:   branchName,
	}
	return nil
}

func GitOpen(dir string) (*Git, error) {
	repo := Git{}
	r, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository: %v", err)
	}
	repo.Repository = r
	return &repo, nil
}

func (repo *Git) getRoot() (string, error) {
	worktree, err := repo.Repository.Worktree()
	if err != nil {
		return "", err
	}
	return worktree.Filesystem.Root(), nil
}

func (repo *Git) TaggedCommit(filename string, message string, tag string, annotatedTag bool, author object.Signature) (*plumbing.Hash, *plumbing.Reference, error) {
	worktree, err := repo.Repository.Worktree()
	if err != nil {
		return nil, nil, err
	}
	root, err := repo.getRoot()
	if err != nil {
		return nil, nil, err
	}
	_, err = os.Create(filepath.Join(root, filename))
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("Add filename to worktree %s", filename)
	_, err = worktree.Add(filename)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Commit %s", message)
	commit1, err := worktree.Commit(message, &git.CommitOptions{Author: &author})
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("tag %s", tag)

	var tagOpts *git.CreateTagOptions = nil
	if annotatedTag {
		tagOpts = &git.CreateTagOptions{Message: tag}
	}

	tag1, err := repo.Repository.CreateTag(tag, commit1, tagOpts)
	if err != nil {
		return nil, nil, err
	}
	return &commit1, tag1, nil
}
