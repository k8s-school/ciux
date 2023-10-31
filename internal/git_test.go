package internal

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

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

func initGitRepo() (*Git, error) {

	gitMeta := Git{}

	dir, err := os.MkdirTemp("/tmp", "ciux-git-test-")
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
	repo, err := initGitRepo()
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
	gitMeta, err := initGitRepo()
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
	gitMeta, err := initGitRepo()
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
	gitMeta, err := initGitRepo()
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
	gitMeta, err := initGitRepo()
	assert.NoError(err)
	root, err := gitMeta.getRoot()
	assert.NoError(err)
	t.Logf("repo root: %s", root)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	assert.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	assert.NoError(err)

	branchName := "testbranch"
	branch := fmt.Sprintf("refs/heads/%s", branchName)
	b := plumbing.ReferenceName(branch)
	err = worktree.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})
	assert.NoError(err)

	assert.True(gitMeta.HasBranch(branchName))
	assert.False(gitMeta.HasBranch("notexist"))

	// Test for remote repository
	gitRemoteMeta, err := GitLsRemote("file://" + worktree.Filesystem.Root())
	assert.NoError(err)
	assert.True(gitRemoteMeta.HasBranch(branchName))
	assert.False(gitRemoteMeta.HasBranch("notexist"))
}
