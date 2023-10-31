package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

var author = object.Signature{
	Name:  "John Doe",
	Email: "test@test.com",
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func gitDescribeTest(assert *assert.Assertions, gitMeta Git, expectedTagName string, expectedCounter int, expectedHeadHash string, expectedDirty bool) {
	err := gitMeta.Describe()
	revision := gitMeta.Revision
	assert.NoError(err)
	assert.Equal(expectedTagName, revision.TagName)
	assert.Equal(expectedCounter, revision.Counter)
	assert.Equal(expectedHeadHash, revision.HeadHash)
	assert.Equal(expectedDirty, revision.Dirty)
}

func initGitRepo(pattern string) (*Git, error) {

	gitMeta := Git{}

	dir := os.TempDir()

	if len(pattern) == 0 {
		pattern = "ciux-git-test-"
	}

	dir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return nil, err
	}
	gitMeta.Repository = repo

	return &gitMeta, nil
}

func TestGitSemverTagMap(t *testing.T) {
	assert := assert.New(t)
	repo, err := initGitRepo("ciux-git-semver-test-")
	assert.NoError(err)

	tags, err := GitSemverTagMap(*repo.Repository)
	assert.NoError(err)
	assert.Equal(map[string]string{}, *tags)

	commit1, _, err := repo.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)

	tags, err = GitSemverTagMap(*repo.Repository)
	assert.NoError(err)
	assert.Equal(map[string]string{
		commit1.String(): "v1.0.0",
	}, *tags)

	commit2, _, err := repo.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	assert.NoError(err)
	tags, err = GitSemverTagMap(*repo.Repository)
	assert.NoError(err)
	t.Log("Comparing tags map")
	assert.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, *tags)

	commit3, tag3, err := repo.TaggedCommit("third.txt", "third", "no-semver", true, author)
	assert.NoError(err)
	assert.NotEqual(commit3.String(), tag3.Hash().String())
	tags, err = GitSemverTagMap(*repo.Repository)
	assert.NoError(err)
	assert.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, *tags)

	commit4, tag4, err := repo.TaggedCommit("fourth.txt", "fourth", "no-annot", false, author)
	assert.NoError(err)
	assert.Equal(commit4.String(), tag4.Hash().String())
	tags, err = GitSemverTagMap(*repo.Repository)
	assert.NoError(err)
	assert.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, *tags)

}

func TestGitDescribe(t *testing.T) {
	assert := assert.New(t)
	gitMeta, err := initGitRepo("ciux-git-describe-test-")
	assert.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	assert.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)
	gitDescribeTest(assert, *gitMeta, "v1.0.0", 0, commit1.String(), false)

	commit2, _ := worktree.Commit("second", &git.CommitOptions{Author: &author})
	gitDescribeTest(assert, *gitMeta, "v1.0.0", 1, commit2.String(), false)

	commit3, _ := worktree.Commit("third", &git.CommitOptions{Author: &author})
	gitDescribeTest(assert, *gitMeta, "v1.0.0", 2, commit3.String(), false)

	// Ignore non annotated tag
	repo.CreateTag("v2.0.0", commit3, nil)
	gitDescribeTest(assert, *gitMeta, "v1.0.0", 2, commit3.String(), false)
}

func TestGitDescribeWithBranch(t *testing.T) {
	assert := assert.New(t)
	gitMeta, err := initGitRepo("ciux-git-describe-test-")
	assert.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	assert.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)
	gitDescribeTest(assert, *gitMeta, "v1.0.0", 0, commit1.String(), false)

	branch := plumbing.ReferenceName("testbranch")
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Create: true,
	})
	assert.NoError(err)

	commit2, _, err := gitMeta.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	assert.NoError(err)
	gitDescribeTest(assert, *gitMeta, "v2.0.0", 0, commit2.String(), false)
}

func TestGitLsRemote(t *testing.T) {
	assert := assert.New(t)
	gitMeta, err := initGitRepo("ciux-git-lsremote-test-")
	assert.NoError(err)
	root, err := gitMeta.getRoot()
	assert.NoError(err)
	t.Logf("repo root: %s", root)
	assert.NoError(err)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	assert.NoError(err)

	// Create branch

	branchName := "testbranch"
	branch := fmt.Sprintf("refs/heads/%s", branchName)
	b := plumbing.ReferenceName(branch)
	err = worktree.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})
	assert.NoError(err)
	branchIter, err := repo.Branches()
	assert.NoError(err)

	branchIter.ForEach(func(r *plumbing.Reference) error {
		t.Logf("Branch %s", r.Name())
		return nil
	})

	gitRemMeta, err := GitLsRemote("file://" + worktree.Filesystem.Root())
	assert.NoError(err)
	// TODO improve this test
	assert.Contains(gitRemMeta.Branches, "master")
	assert.Contains(gitRemMeta.Branches, "testbranch")
}

func TestHasBranch(t *testing.T) {
	assert := assert.New(t)

	// Test for local repository
	gitMeta, err := initGitRepo("ciux-git-hasbranch-test-")
	assert.NoError(err)
	root, err := gitMeta.getRoot()
	assert.NoError(err)
	t.Logf("repo root: %s", root)
	defer os.RemoveAll(root)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	assert.NoError(err)

	branchName := "testbranch"
	err = gitMeta.CreateBranch(branchName)
	assert.NoError(err)

	assert.True(gitMeta.HasBranch(branchName))
	assert.False(gitMeta.HasBranch("notexist"))

	// Test for remote repository
	gitRemoteMeta, err := GitLsRemote("file://" + worktree.Filesystem.Root())
	assert.NoError(err)
	assert.True(gitRemoteMeta.HasBranch(branchName))
	assert.False(gitRemoteMeta.HasBranch("notexist"))
}
func TestGetDepsBranches(t *testing.T) {
	assert := assert.New(t)

	// Create a temporary directory for the git repository
	gitMeta, err := initGitRepo("ciux-git-getversions-test-")
	assert.NoError(err)
	root, err := gitMeta.getRoot()
	assert.NoError(err)
	defer os.RemoveAll(root)

	// Create a custom .ciux file in the repository
	pattern := "ciux-git-getversions-test-"
	gitDepMeta, err := initGitRepo(pattern)
	assert.NoError(err)
	depRoot, err := gitDepMeta.getRoot()
	defer os.RemoveAll(depRoot)
	assert.NoError(err)
	dep := Dependency{
		Url:   "file://" + depRoot,
		Clone: true,
		Pull:  true,
	}
	c := Config{
		Registry:     "test-registry.io",
		Dependencies: []Dependency{dep},
	}

	items := map[string]interface{}{}
	err = mapstructure.Decode(c, &items)
	assert.NoError(err)
	err = viper.MergeConfigMap(items)
	assert.NoError(err)

	yamlData, err := yaml.Marshal(items)
	assert.NoError(err)
	t.Logf("yamlData: %s", string(yamlData))
	ciuxPath := filepath.Join(root, ".ciux")
	f, err := os.Create(ciuxPath)
	assert.NoError(err)
	_, err = f.Write(yamlData)
	assert.NoError(err)
	f.Close()

	// Add the file to the repository
	worktree, err := gitMeta.Repository.Worktree()
	assert.NoError(err)
	_, err = worktree.Add(".ciux")
	assert.NoError(err)

	// Commit the changes to the repository
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{})
	assert.NoError(err)

	// Create a new tag for the commit
	_, err = gitMeta.Repository.CreateTag("v1.0.0", commit, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Message: "Test tag",
	})
	assert.NoError(err)

	// Initialize the dependency repository
	_, _, err = gitDepMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)

	depGits, err := GetDepsBranch(root)
	assert.NoError(err)

	// Assert that the dependency has the correct branch information
	assert.Equal("master", (*depGits)[0].Revision.Branch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = gitMeta.CreateBranch(branchName)
	assert.NoError(err)

	depGits, err = GetDepsBranch(root)
	assert.NoError(err)
	assert.Equal("master", (*depGits)[0].Revision.Branch)

	// Create testbranch in the dependency repository
	err = gitDepMeta.CreateBranch(branchName)
	assert.NoError(err)

	depGits, err = GetDepsBranch(root)
	assert.NoError(err)
	assert.Equal("testbranch", (*depGits)[0].Revision.Branch)

}
