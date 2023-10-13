package internal

import (
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

// GitTagMap ...
func GitTagMap(repo git.Repository) (*map[string]string, error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	tagMap := map[string]string{}
	err = iter.ForEach(func(r *plumbing.Reference) error {
		tag, _ := repo.TagObject(r.Hash())
		if SemVerParse(r.Name().Short()) == nil {
			// Filter out tags that are not semver
			return nil
		}
		if tag == nil {
			tagMap[r.Hash().String()] = r.Name().Short()
		} else {
			c, err := tag.Commit()
			if err != nil {
				return err
			}
			tagMap[c.Hash.String()] = r.Name().Short()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &tagMap, nil
}

func GitBranchName(repo git.Repository) (*string, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("unable to find head: %v", err)
	}
	branchName := head.Name().Short()
	return &branchName, nil
}

// GitLsRemote
// https://github.com/go-git/go-git/blob/master/_examples/ls-remote/main.go
func GitLsRemote(repoUrl string, reference string) {

	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoUrl},
	})

	refs, err := rem.List(&git.ListOptions{
		// Returns all references, including peeled references.
		PeelingOption: git.AppendPeeled,
	})

	if err != nil {
		// TODO share logger between packages
		log.Fatal(err)
	}

	// Find annotated tags
	// the one with ^{} is the annotated tag
	// 2023/10/13 18:18:06 Tags found: v0.1
	// 2023/10/13 18:18:06 Tags found: lsst-france-meeting-may-2014
	// 2023/10/13 18:18:06 Tags found: test
	// 2023/10/13 18:18:06 Tags found: v0.5
	// 2023/10/13 18:18:06 Tags found: v0.1^{}
	// 2023/10/13 18:18:06 Tags found: v0.5^{}

	var tags []string
	for _, ref := range refs {
		if ref.Name().IsTag() {
			log.Printf("Tags found: %v", ref.Name().Short())
			tags = append(tags, ref.Name().Short())
		}
	}

	if len(tags) == 0 {
		log.Println("No tags!")
	}

	log.Printf("Tags found: %v", tags)

}

// GitDescribe
func GitDescribe(repo git.Repository) (*string, *int, *string, *bool, error) {
	type gitDescribeNode struct {
		Commit   object.Commit
		Distance int
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to find worktree: %v", err)
	}
	status, err := w.Status()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to find worktree status: %v", err)
	}

	var dirty bool
	if status.IsClean() {
		dirty = false
	} else {
		dirty = true
	}

	head, err := repo.Head()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to find head: %v", err)
	}
	headHash := head.Hash().String()
	tags, err := GitTagMap(repo)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to get tags: %v", err)
	}
	commits, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to get log: %v", err)
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
	if tagHash == "" {
		for _, node := range state {
			if node.Distance+1 > counter {
				counter = node.Distance + 1
			}
		}
		tagName := ""
		return &tagName, &counter, &headHash, &dirty, nil
	}
	tagName := (*tags)[tagHash]
	return &tagName, &counter, &headHash, &dirty, nil
}
