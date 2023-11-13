package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

var author = object.Signature{
	Name:  "John Doe",
	Email: "test@test.com",
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func gitDescribeTest(require *require.Assertions, gitMeta Git, expectedTagName string, expectedCounter int, expectedHeadHash string, expectedDirty bool) {
	revision, err := gitMeta.GetRevision()
	require.NoError(err)
	require.Equal(expectedTagName, revision.Tag)
	require.Equal(expectedCounter, revision.Counter)
	require.Equal(expectedHeadHash, revision.HeadHash)
	require.Equal(expectedDirty, revision.Dirty)
}

func initGitRepo(pattern string) (Git, error) {
	gitObj := Git{}
	dir, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return gitObj, err
	}
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return gitObj, err
	}
	gitObj.Repository = repo

	return gitObj, nil
}
func TestGitSemverTagMap(t *testing.T) {
	require := require.New(t)
	git, err := initGitRepo("ciux-git-semver-test-")
	require.NoError(err)

	tags, err := GitSemverTagMap(*git.Repository)
	require.NoError(err)
	require.Equal(map[string]string{}, FormatTags(tags))

	commit1, _, err := git.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	tags, err = GitSemverTagMap(*git.Repository)
	require.NoError(err)
	require.Equal(map[string]string{
		commit1.String(): "v1.0.0",
	}, FormatTags(tags))

	commit2, _, err := git.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)
	tags, err = GitSemverTagMap(*git.Repository)
	require.NoError(err)
	t.Log("Comparing tags map")
	require.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, FormatTags(tags))

	commit3, tag3, err := git.TaggedCommit("third.txt", "third", "no-semver", true, author)
	require.NoError(err)
	require.NotEqual(commit3.String(), tag3.Hash().String())
	tags, err = GitSemverTagMap(*git.Repository)
	require.NoError(err)
	require.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, FormatTags(tags))

	commit4, tag4, err := git.TaggedCommit("fourth.txt", "fourth", "no-annot", false, author)
	require.NoError(err)
	require.Equal(commit4.String(), tag4.Hash().String())
	tags, err = GitSemverTagMap(*git.Repository)
	require.NoError(err)
	require.Equal(map[string]string{
		commit1.String(): "v1.0.0",
		commit2.String(): "v2.0.0",
	}, FormatTags(tags))

}

func TestGitDescribe(t *testing.T) {
	require := require.New(t)
	gitMeta, err := initGitRepo("ciux-git-describe-test-")
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	gitDescribeTest(require, gitMeta, "v1.0.0", 0, commit1.String(), false)

	commit2, _ := worktree.Commit("second", &git.CommitOptions{Author: &author})
	gitDescribeTest(require, gitMeta, "v1.0.0", 1, commit2.String(), false)

	commit3, _ := worktree.Commit("third", &git.CommitOptions{Author: &author})
	gitDescribeTest(require, gitMeta, "v1.0.0", 2, commit3.String(), false)

	// Ignore non annotated tag
	repo.CreateTag("v2.0.0", commit3, nil)
	gitDescribeTest(require, gitMeta, "v1.0.0", 2, commit3.String(), false)
}

func TestGitDescribeWithBranch(t *testing.T) {
	require := require.New(t)
	gitMeta, err := initGitRepo("ciux-git-describe-test-")
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	gitDescribeTest(require, gitMeta, "v1.0.0", 0, commit1.String(), false)

	branch := plumbing.ReferenceName("testbranch")
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Create: true,
	})
	require.NoError(err)

	commit2, _, err := gitMeta.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)
	gitDescribeTest(require, gitMeta, "v2.0.0", 0, commit2.String(), false)
}

func TestGitLsRemote(t *testing.T) {
	require := require.New(t)
	gitMeta, err := initGitRepo("ciux-git-lsremote-test-")
	require.NoError(err)
	root, err := gitMeta.GetRoot()
	require.NoError(err)
	t.Logf("repo root: %s", root)
	require.NoError(err)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	// Create branch

	branchName := "testbranch"
	branch := fmt.Sprintf("refs/heads/%s", branchName)
	b := plumbing.ReferenceName(branch)
	err = worktree.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})
	require.NoError(err)
	branchIter, err := repo.Branches()
	require.NoError(err)

	branchIter.ForEach(func(r *plumbing.Reference) error {
		t.Logf("Branch %s", r.Name())
		return nil
	})

	gitRemMeta, err := GitLsRemote("file://" + worktree.Filesystem.Root())
	require.NoError(err)
	// TODO improve this test
	require.Contains(gitRemMeta.Branches, "master")
	require.Contains(gitRemMeta.Branches, "testbranch")
}

func TestHasBranch(t *testing.T) {
	require := require.New(t)

	// Test for local repository
	gitMeta, err := initGitRepo("ciux-git-hasbranch-test-")
	require.NoError(err)
	root, err := gitMeta.GetRoot()
	require.NoError(err)
	t.Logf("repo root: %s", root)
	defer os.RemoveAll(root)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	require.True(gitMeta.HasBranch("master"))

	branchName := "testbranch"
	err = gitMeta.CreateBranch(branchName)
	require.NoError(err)

	require.True(gitMeta.HasBranch(branchName))
	require.False(gitMeta.HasBranch("notexist"))

	// Test for remote repository
	gitRemoteMeta, err := GitLsRemote("file://" + worktree.Filesystem.Root())
	require.NoError(err)
	require.True(gitRemoteMeta.HasBranch(branchName))
	require.False(gitRemoteMeta.HasBranch("notexist"))
}
func TestCloneWorkBranch(t *testing.T) {
	require := require.New(t)

	// Test for local repository
	gitMeta, err := initGitRepo("ciux-git-clonebranch-test-")
	require.NoError(err)
	root, err := gitMeta.GetRoot()
	require.NoError(err)
	t.Logf("repo root: %s", root)
	defer os.RemoveAll(root)
	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	branchName := "testbranch"
	err = gitMeta.CreateBranch(branchName)
	require.NoError(err)

	commit2, _, err := gitMeta.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)

	require.True(gitMeta.HasBranch(branchName))
	require.False(gitMeta.HasBranch("notexist"))

	// Create a new Git object with the URL and branch of the repository

	gitObj := &Git{
		Url:        "file://" + root,
		WorkBranch: branchName,
	}

	err = gitObj.CloneOrOpen("", true)
	require.NoError(err)
	cloneRoot, err := gitObj.GetRoot()
	require.NoError(err)
	t.Logf("clone repo root: %s", root)
	defer os.RemoveAll(cloneRoot)
	require.NoError(err)

	// Check that the cloned repository has the correct branch checked out
	cloneHead, err := gitObj.Repository.Head()
	require.NoError(err)
	require.Equal(branchName, cloneHead.Name().Short())

	gitObj.GetRevision()
	gitDescribeTest(require, gitMeta, "v2.0.0", 0, commit2.String(), false)

}
func TestGetVersion(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name     string
		rev      GitRevision
		expected string
	}{
		{
			name: "simple",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  1,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expected: "v1.0.0-1-g1234567",
		},
		{
			name: "dirty",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  1,
				HeadHash: "1234567890abcdef",
				Dirty:    true,
			},
			expected: "v1.0.0-1-g1234567-dirty",
		},
		{
			name: "tag",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  0,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expected: "v1.0.0",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			actual := tt.rev.GetVersion()
			require.Equal(tt.expected, actual)
		})
	}

}
func TestMainBranch(t *testing.T) {
	require := require.New(t)

	var mainBranch string
	gitLocal, err := initGitRepo("ciux-git-mainbranch-test-")
	require.NoError(err)
	_, _, err = gitLocal.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	mainBranch, err = gitLocal.MainBranch()
	require.NoError(err)
	require.Equal("master", mainBranch)

	tests := []struct {
		name     string
		url      string
		clone    bool
		expected string
	}{
		{
			name:     "master",
			url:      "https://github.com/astrolabsoftware/fink-alert-simulator",
			clone:    true,
			expected: "master",
		},
		{
			name:     "main",
			url:      "https://github.com/astrolabsoftware/finkctl",
			clone:    false,
			expected: "main",
		},
	}

	for _, tt := range tests {
		gitObj := &Git{
			Url: tt.url,
		}
		if tt.clone {
			err := gitObj.CloneOrOpen("", false)
			require.NoError(err)
		}
		mainBranch, err = gitObj.MainBranch()
		require.NoError(err)
		require.Equal(tt.expected, mainBranch)
	}

}
func TestGetName(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple",
			url:      "https://github.com/astrolabsoftware/fink-alert-simulator",
			expected: "fink-alert-simulator",
		},
		{
			name:     "with-suffix",
			url:      "https://github.com/astrolabsoftware/fink-alert-simulator.git",
			expected: "fink-alert-simulator",
		},
	}

	for _, tt := range tests {
		gitObj := &Git{
			Url: tt.url,
		}
		actual, err := gitObj.GetName()
		require.NoError(err)
		require.Equal(tt.expected, actual)
	}
}
func TestIsDirty(t *testing.T) {
	require := require.New(t)

	// Create a new Git repository
	gitObj, err := initGitRepo("ciux-git-test-")
	require.NoError(err)
	root, err := gitObj.GetRoot()
	require.NoError(err)
	defer os.Remove(root)

	worktree, err := gitObj.Repository.Worktree()
	require.NoError(err)

	// Check that the repository is not dirty
	status, err := worktree.Status()
	require.NoError(err)
	require.False(IsDirty(status))

	// Create a new untracked file and check that the repository is not dirty

	fname := "new-file"
	_, err = os.Create(filepath.Join(root, fname))
	require.NoError(err)
	require.NoError(err)
	status, err = worktree.Status()
	require.NoError(err)
	require.False(IsDirty(status))

	// Stage the file and check that the repository is dirty
	_, err = worktree.Add("new-file")
	require.NoError(err)
	status, err = worktree.Status()
	require.NoError(err)
	require.True(IsDirty(status))

}
