package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/k8s-school/ciux/log"
)

type GitRevision struct {
	Tag      string
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
	WorkBranch string
	Author     string
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

// FormatTags convert a map of tags to a map of human readable tags
func FormatTags(tags *map[plumbing.Hash]*plumbing.Reference) map[string]string {
	tagMap := map[string]string{}
	for key, ref := range *tags {
		tagMap[key.String()] = ref.Name().Short()
	}
	return tagMap
}

func (gitObj *Git) MainBranch() (string, error) {

	mainBranch := ""
	found := false
	mainNames := []string{"main", "master"}

	name, err := gitObj.GetName()
	if err != nil {
		return "", fmt.Errorf("unable to get name for git repository %s: %v", gitObj.Url, err)
	}

	for _, branch := range mainNames {

		found, err = gitObj.HasBranch(branch)
		if err != nil {
			return "", fmt.Errorf("unable to look for main branch: %v", err)
		}
		if found {
			mainBranch = branch
			break
		}
	}

	if !found {
		return "", fmt.Errorf("unable to find main branch for git repository %s", name)
	}
	return mainBranch, nil
}

func (gitObj *Git) GetName() (string, error) {
	var lastDir string
	var err error
	if len(gitObj.Url) != 0 {
		var trimmedUrl string = strings.TrimRight(gitObj.Url, ".git")
		lastDir, err = LastDir(trimmedUrl)
		if err != nil {
			return "", err
		}
	} else {
		root, err := gitObj.GetRoot()
		if err != nil {
			return "", err
		}
		lastDir, err = LastDir(root)
		if err != nil {
			return "", err
		}
	}
	return lastDir, nil
}

func (gitObj *Git) GetEnVarName() (string, error) {
	varName, err := gitObj.GetName()
	if err != nil {
		return varName, fmt.Errorf("unable to get name for git repository %v: %v", gitObj, err)
	}
	varName = strings.ReplaceAll(varName, "-", "_")
	varName = strings.ToUpper(varName)
	return varName, nil
}

func (gitObj *Git) Clone(basePath string, singleBranch bool) error {
	name, err := gitObj.GetName()
	if err != nil {
		return fmt.Errorf("unable to get name from url %s: %v", gitObj.Url, err)
	}
	var repoDir string
	if basePath == "" {
		repoDir, err = os.MkdirTemp(os.TempDir(), "ciux-"+name+"-")
		if err != nil {
			return err
		}
	} else {
		repoDir = filepath.Join(basePath, name)
		err := os.MkdirAll(repoDir, 0755)
		if err != nil {
			return err
		}
	}
	var refName plumbing.ReferenceName
	if singleBranch {
		refName = plumbing.ReferenceName(gitObj.WorkBranch)
	}
	repository, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           gitObj.Url,
		ReferenceName: refName,
		SingleBranch:  singleBranch,
		Progress:      os.Stdout,
	})
	if err == git.ErrRepositoryAlreadyExists {
		log.Warnf("not cloning dependency repository %s, working with existing one: %s", gitObj.Url, repoDir)
		repository, err = git.PlainOpen(repoDir)
		if err != nil {
			return fmt.Errorf("unable to open git repository %s: %v", gitObj.Url, err)
		}
	} else if err != nil {
		return fmt.Errorf("unable to clone git repository %s: %v", gitObj.Url, err)
	}
	gitObj.Repository = repository
	return nil
}

// GitLsRemote returns branches and tag of a remote repository
// https://github.com/go-git/go-git/blob/master/_examples/ls-remote/main.go
func GitLsRemote(url string) (repo *Git, err error) {
	repo = &Git{Url: url}

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

func (gitObj *Git) isRemote() bool {
	return gitObj.Repository == nil
}

func (gitObj *Git) HasBranch(branchname string) (bool, error) {
	found := false
	if gitObj.isRemote() {
		remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
			Name: "origin",
			URLs: []string{gitObj.Url},
		})
		refs, err := remote.List(&git.ListOptions{
			// Returns all references, including peeled references.
			PeelingOption: git.AppendPeeled,
		})
		if err != nil {
			return false, fmt.Errorf("unable to list remote references: %+v", err)
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
		bIter, err := gitObj.Repository.Branches()
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

// IsDirty returns true if all the files are in Unmodified or Untracked status.
func IsDirty(s git.Status) bool {
	for _, status := range s {
		if status.Worktree == git.Untracked {
			continue
		} else if status.Worktree != git.Unmodified || status.Staging != git.Unmodified {
			return true
		}
	}

	return false
}

// GetRevision the reference as 'git describe ' will do
func (g *Git) GetRevision() (*GitRevision, error) {

	head, err := g.Repository.Head()
	if err != nil {
		return nil, fmt.Errorf("unable to find head: %v", err)
	}
	w, err := g.Repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("unable to find worktree: %v", err)
	}
	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("unable to find worktree status: %v", err)
	}
	branchName := head.Name().Short()
	headHash := head.Hash().String()
	dirty := IsDirty(status)

	// Fetch the reference log
	cIter, err := g.Repository.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get reference log: %v", err)
	}

	// Build the semver annotated tag map
	semverTags, err := GitSemverTagMap(*g.Repository)
	if err != nil {
		return nil, fmt.Errorf("unable to get semver tags: %v", err)
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
		return nil, fmt.Errorf("unable to loop on commits: %v", err)
	}
	rev := GitRevision{
		Tag:      tag.Name().Short(),
		Counter:  count,
		HeadHash: headHash,
		Dirty:    dirty,
		Branch:   branchName,
	}
	return &rev, nil
}

// GetVersion returns the reference as 'git describe ' will do
// except that tag is the latest semver annotated tag
func (rev *GitRevision) GetVersion() string {
	var dirty string
	if rev.Dirty {
		dirty = "-dirty"
	}
	var counterHash string
	if rev.Counter != 0 {
		counterHash = fmt.Sprintf("-%d-g%s", rev.Counter, rev.HeadHash[0:7])
	}
	version := fmt.Sprintf("%s%s%s", rev.Tag, counterHash, dirty)
	return version
}

func NewGit(dir string) (Git, error) {
	repo := Git{}
	r, err := git.PlainOpen(dir)
	if err != nil {
		return repo, fmt.Errorf("unable to open git repository: %v", err)
	}
	repo.Repository = r
	return repo, nil
}

func (git *Git) GetRoot() (string, error) {
	if git.Repository == nil {
		return "", fmt.Errorf("repository is nil for git %+v", git)
	}
	worktree, err := git.Repository.Worktree()
	if err != nil {
		return "", err
	}
	return worktree.Filesystem.Root(), nil
}

func (repo *Git) CreateBranch(branchName string) error {
	worktree, err := repo.Repository.Worktree()
	if err != nil {
		return err
	}

	branch := fmt.Sprintf("refs/heads/%s", branchName)
	b := plumbing.ReferenceName(branch)
	err = worktree.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})
	if err != nil {
		return err
	}
	return nil
}

func (repo *Git) TaggedCommit(filename string, message string, tag string, annotatedTag bool, author object.Signature) (*plumbing.Hash, *plumbing.Reference, error) {
	worktree, err := repo.Repository.Worktree()
	if err != nil {
		return nil, nil, err
	}
	root, err := repo.GetRoot()
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
