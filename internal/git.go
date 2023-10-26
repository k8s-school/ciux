package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

type GitMeta struct {
	Tags       []string
	Branches   []string
	Url        string
	Repository *git.Repository
	Revision   GitRevision
	Author     string
	Directory  string
}

// GitSemverTagMap ...
func GitSemverTagMap(repo git.Repository) (*map[string]string, error) {
	tagIter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	tagMap := map[string]string{}
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
			tagMap[c.Hash.String()] = r.Name().Short()
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

// GitLsRemote
// https://github.com/go-git/go-git/blob/master/_examples/ls-remote/main.go
func GitLsRemote(url string) (repo *GitMeta, err error) {
	repo = &GitMeta{}
	repo.Url = url

	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo.Url},
	})

	refs, err := rem.List(&git.ListOptions{
		// Returns all references, including peeled references.
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		// TODO share logger between packages
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

// GitDescribe
func (repo *GitMeta) GitDescribe() error {
	type gitDescribeNode struct {
		Commit   object.Commit
		Distance int
	}

	w, err := repo.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("unable to find worktree: %v", err)
	}
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("unable to find worktree status: %v", err)
	}

	var dirty bool
	if status.IsClean() {
		dirty = false
	} else {
		dirty = true
	}

	head, err := repo.Repository.Head()
	if err != nil {
		return fmt.Errorf("unable to find head: %v", err)
	}
	headHash := head.Hash().String()
	tags, err := GitSemverTagMap(*repo.Repository)
	if err != nil {
		return fmt.Errorf("unable to get tags: %v", err)
	}
	commits, err := repo.Repository.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return fmt.Errorf("unable to get log: %v", err)
	}
	state := map[string]gitDescribeNode{}
	counter := 0
	tagHash := ""
	commits.ForEach(func(c *object.Commit) error {
		node, found := state[c.Hash.String()]
		if !found {
			node = gitDescribeNode{
				Commit:   *c,
				Distance: 0,
			}
			state[c.Hash.String()] = node
		}
		c.Parents().ForEach(func(p *object.Commit) error {
			_, found := state[p.Hash.String()]
			if !found {
				state[p.Hash.String()] = gitDescribeNode{
					Commit:   *p,
					Distance: node.Distance + 1,
				}
			}
			return nil
		})

		_, foundTag := (*tags)[c.Hash.String()]
		if tagHash == "" && foundTag {
			counter = state[c.Hash.String()].Distance
			tagHash = c.Hash.String()
		}
		return nil
	})
	var tagName string
	if tagHash == "" {
		for _, node := range state {
			if node.Distance+1 > counter {
				counter = node.Distance + 1
			}
		}
		tagName = ""
	} else {
		tagName = (*tags)[tagHash]
	}

	branchName := head.Name().Short()

	repo.Revision = GitRevision{
		TagName:  tagName,
		Counter:  counter,
		HeadHash: headHash,
		Dirty:    dirty,
		Branch:   branchName,
	}
	return nil
}

func GetRevision(dir string, deps []string) (repo *GitMeta, e error) {
	repo = &GitMeta{}
	r, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository: %v", err)
	}
	repo.Repository = r
	err = repo.GitDescribe()
	if err != nil {
		return nil, fmt.Errorf("unable to describe commit: %v", err)
	}

	log.Debugf("%+v", repo.Revision)

	return repo, nil
}

func (repo *GitMeta) TaggedCommit(filename string, message string, tag string, annotatedTag bool, author object.Signature) (*plumbing.Hash, *plumbing.Reference, error) {
	worktree, err := repo.Repository.Worktree()
	if err != nil {
		return nil, nil, err
	}

	_, err = os.Create(filepath.Join(worktree.Filesystem.Root(), filename))
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
