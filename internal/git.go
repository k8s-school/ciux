package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/k8s-school/ciux/log"
)

type Git struct {
	InPlace        bool
	RemoteTags     []string
	RemoteBranches []string
	Url            string
	Repository     *git.Repository
	WorkBranch     string
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

func (gitObj *Git) GetEnVarPrefix() (string, error) {
	varName, err := gitObj.GetName()
	if err != nil {
		return varName, fmt.Errorf("unable to get name for git repository %v: %v", gitObj, err)
	}
	varName = strings.ReplaceAll(varName, "-", "_")
	varName = strings.ToUpper(varName)
	return varName, nil
}

func (gitObj *Git) OpenIfExists(destBasePath string) error {
	name, err := gitObj.GetName()
	if err != nil {
		return fmt.Errorf("unable to get name from url %s: %v", gitObj.Url, err)
	}
	destPath := filepath.Join(destBasePath, name)
	// Check that destPath is a directory

	_, err = os.Stat(destPath)
	var repository *git.Repository
	if !os.IsNotExist(err) {
		repository, err = git.PlainOpen(destPath)
		if err != nil {
			return fmt.Errorf("unable to open git repository %s: %v", gitObj.Url, err)
		}
	} else {
		slog.Debug("Repository does not exist locally", "url", gitObj.Url, "path", destPath)
	}
	gitObj.Repository = repository
	return nil
}

func (gitObj *Git) CloneOrOpen(destBasePath string, singleBranch bool) error {
	name, err := gitObj.GetName()
	if err != nil {
		return fmt.Errorf("unable to get name from url %s: %v", gitObj.Url, err)
	}
	var destPath string
	if destBasePath == "" {
		destPath, err = os.MkdirTemp(os.TempDir(), "ciux-"+name+"-")
		if err != nil {
			return err
		}
	} else {
		destPath = filepath.Join(destBasePath, name)
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			return err
		}
	}
	var refName plumbing.ReferenceName
	if singleBranch {
		refName = plumbing.ReferenceName(gitObj.WorkBranch)
	}
	var progress sideband.Progress = nil
	if log.IsDebugEnabled() {
		progress = os.Stdout
	}
	options := &git.CloneOptions{
		URL:           gitObj.Url,
		ReferenceName: refName,
		SingleBranch:  singleBranch,
		Progress:      progress,
	}
	// TODO check if repository already exists, then try to open it else clone it
	repository, err := git.PlainClone(destPath, false, options)
	if err == git.ErrRepositoryAlreadyExists {
		gitObj.InPlace = true
		slog.Warn("In place repository", "url", gitObj.Url, "path", destPath)
		repository, err = git.PlainOpen(destPath)
		if err != nil {
			return fmt.Errorf("unable to open git repository %s: %v", gitObj.Url, err)
		}
	} else if err != nil {
		return fmt.Errorf("unable to clone git repository %s: %v", gitObj.Url, err)
	}
	gitObj.Repository = repository
	return nil
}

// LsRemote returns branches and tag of a remote repository
// https://github.com/go-git/go-git/blob/master/_examples/ls-remote/main.go
func (gitObj *Git) LsRemote() error {

	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitObj.Url},
	})

	refs, err := remote.List(&git.ListOptions{
		// Returns all references, including peeled references.
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return fmt.Errorf("unable to list remote references: %v", err)
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
			gitObj.RemoteBranches = append(gitObj.RemoteBranches, ref.Name().Short())
		}
		if ref.Name().IsTag() {
			gitObj.RemoteTags = append(gitObj.RemoteTags, ref.Name().Short())
		}
	}
	return nil
}

// isRemoteOnly returns true if the git object is only a remote repository
// i.e. it has not been cloned locally
func (gitObj *Git) isRemoteOnly() bool {
	return gitObj.Repository == nil
}

func (gitObj *Git) HasBranch(branchname string) (bool, error) {
	found := false
	if gitObj.isRemoteOnly() {
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
	slog.Debug("Branch search", "url", gitObj.Url, "branchname", branchname, "found", found)
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

// GetRevision returns the reference as 'git checkout <hash> && git describe ' would do
func (g *Git) GetRevision(hash plumbing.Hash) (*GitRevision, error) {

	w, err := g.Repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("unable to find worktree: %v", err)
	}
	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("unable to find worktree status: %v", err)
	}
	dirty := IsDirty(status)

	// Fetch the reference log
	cIter, err := g.Repository.Log(&git.LogOptions{
		From:  hash,
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
	var tagStr string
	if tag == nil {
		tagStr = ""
	} else {
		tagStr = tag.Name().Short()
	}
	rev := GitRevision{
		Tag:     tagStr,
		Counter: count,
		Hash:    hash.String(),
		Dirty:   dirty,
	}
	return &rev, nil
}

// GetHeadRevision the reference as 'git describe ' will do
func (g *Git) GetHeadRevision() (*GitRevision, error) {

	head, err := g.Repository.Head()
	if err != nil {
		return nil, fmt.Errorf("unable to find head: %v", err)
	}

	branchName := head.Name().Short()
	revision, err := g.GetRevision(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("unable to get head revision: %v", err)
	}

	revision.Branch = branchName
	return revision, nil
}

func (g *Git) GetBranch() (string, error) {
	head, err := g.Repository.Head()
	if err != nil {
		return "", fmt.Errorf("unable to find head: %v", err)
	}
	return head.Name().Short(), nil
}

func NewGit(dir string) (*Git, error) {
	repo := Git{}
	r, err := git.PlainOpen(dir)
	if err != nil {
		return &repo, fmt.Errorf("unable to open git repository: %v", err)
	}
	repo.Repository = r
	return &repo, nil
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

func (git *Git) IsGoModule() (bool, error) {
	root, err := git.GetRoot()
	if err != nil {
		return false, fmt.Errorf("unable to get root of git repository: %v", err)
	}

	// Check if go.mod and go.sum exist in root
	for _, file := range []string{"go.mod", "go.sum"} {
		modFile := filepath.Join(root, file)
		_, err = os.Stat(modFile)
		if os.IsNotExist(err) {
			slog.Debug("Not a go module", "dependency_path", root)
			return false, nil
		}
	}
	return true, nil
}

func (git *Git) GoInstall() error {
	root, err := git.GetRoot()
	if err != nil {
		return fmt.Errorf("unable to get root of git repository: %v", err)
	}

	cmd := fmt.Sprintf("go install -C %s", root)
	outstr, errstr, err := ExecCmd(cmd, false, false)
	slog.Debug("Install from source", "cmd", cmd, "out", outstr, "err", errstr)

	if err != nil {
		return fmt.Errorf("unable to install go modules for git repository %s: %v", git.Url, err)
	}
	return nil
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
	slog.Debug("Add filename to worktree", "file", filename)
	_, err = worktree.Add(filename)
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("Commit", "message", message)
	commit1, err := worktree.Commit(message, &git.CommitOptions{Author: &author})
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("Create tag", "tag", tag)

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
