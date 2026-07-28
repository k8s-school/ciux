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

func getHeadRevisionTest(require *require.Assertions, gitMeta Git, expectedTagName string, expectedCounter int, expectedHeadHash string, expectedDirty bool) {
	revision, err := gitMeta.GetHeadRevision()
	require.NoError(err)
	require.Equal(expectedTagName, revision.Tag)
	require.Equal(expectedCounter, revision.Counter)
	require.Equal(expectedHeadHash, revision.Hash)
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

func TestGetHeadRevision(t *testing.T) {
	require := require.New(t)
	gitMeta, err := initGitRepo("ciux-git-getrevision-test-")
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	getHeadRevisionTest(require, gitMeta, "v1.0.0", 0, commit1.String(), false)

	// Counting commits since the tag is what matters here, their content does
	// not: go-git rejects empty commits unless allowed explicitly
	commit2, err := worktree.Commit("second", &git.CommitOptions{Author: &author, AllowEmptyCommits: true})
	require.NoError(err)
	getHeadRevisionTest(require, gitMeta, "v1.0.0", 1, commit2.String(), false)

	commit3, err := worktree.Commit("third", &git.CommitOptions{Author: &author, AllowEmptyCommits: true})
	require.NoError(err)
	getHeadRevisionTest(require, gitMeta, "v1.0.0", 2, commit3.String(), false)

	// Ignore non annotated tag
	repo.CreateTag("v2.0.0", commit3, nil)
	getHeadRevisionTest(require, gitMeta, "v1.0.0", 2, commit3.String(), false)
}

func TestGetHeadRevisionWithBranch(t *testing.T) {
	require := require.New(t)
	gitMeta, err := initGitRepo("ciux-git-getrevision-branch-test-")
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	commit1, _, err := gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	getHeadRevisionTest(require, gitMeta, "v1.0.0", 0, commit1.String(), false)

	branchName := "testbranch"
	branch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Create: true,
	})
	require.NoError(err)

	commit2, _, err := gitMeta.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)
	getHeadRevisionTest(require, gitMeta, "v2.0.0", 0, commit2.String(), false)
	rev, err := gitMeta.GetHeadRevision()
	require.NoError(err)

	require.Equal(branchName, rev.Branch)

}

func TestGetBranch(t *testing.T) {
	require := require.New(t)

	gitMeta, err := initGitRepo("ciux-git-getbranch-test-")
	require.NoError(err)
	repo := gitMeta.Repository
	worktree, err := repo.Worktree()
	require.NoError(err)

	_, _, err = gitMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	branchName := "testbranch"
	branch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Create: true,
	})
	require.NoError(err)

	actualBranch, err := gitMeta.GetBranch()
	require.NoError(err)
	require.Equal(branchName, actualBranch)
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

	gitRemoteMeta := Git{
		Url: "file://" + worktree.Filesystem.Root(),
	}
	err = gitRemoteMeta.LsRemote()
	require.NoError(err)
	require.NoError(err)
	// TODO improve this test
	require.Contains(gitRemoteMeta.RemoteBranches, "master")
	require.Contains(gitRemoteMeta.RemoteBranches, "testbranch")
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
	gitRemoteMeta := Git{
		Url: "file://" + worktree.Filesystem.Root(),
	}
	err = gitRemoteMeta.LsRemote()
	require.NoError(err)
	require.True(gitRemoteMeta.HasBranch(branchName))
	require.False(gitRemoteMeta.HasBranch("notexist"))
}
func TestCloneWorkBranch(t *testing.T) {
	require := require.New(t)

	// Test for local repository
	gitOrigin, err := initGitRepo("ciux-git-clonebranch-test-")
	require.NoError(err)
	rootOrigin, err := gitOrigin.GetRoot()
	require.NoError(err)
	t.Logf("repo root: %s", rootOrigin)
	_, _, err = gitOrigin.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	branchName := "testbranch"
	err = gitOrigin.CreateBranch(branchName)
	require.NoError(err)

	commit2, _, err := gitOrigin.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)

	require.True(gitOrigin.HasBranch(branchName))
	require.False(gitOrigin.HasBranch("notexist"))

	// Create a new Git object with the URL and branch of the repository

	gitObj := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: branchName,
	}

	err = gitObj.CloneOrOpen("", true, false)
	require.NoError(err)
	cloneRoot, err := gitObj.GetRoot()
	require.NoError(err)
	t.Logf("clone repo root: %s", rootOrigin)

	// Check that the cloned repository has the correct branch checked out
	cloneHead, err := gitObj.Repository.Head()
	require.NoError(err)
	require.Equal(branchName, cloneHead.Name().Short())

	gitObj.GetHeadRevision()
	getHeadRevisionTest(require, gitOrigin, "v2.0.0", 0, commit2.String(), false)

	os.RemoveAll(rootOrigin)
	os.RemoveAll(cloneRoot)
}

// A dependency already cloned in the base path is used as is, whatever branch
// it is on: this is what pins a dependency to a stale revision on a CI runner
// which reuses its workspace. Refreshing clones it again on the work branch.
func TestCloneOrOpenRefresh(t *testing.T) {
	require := require.New(t)

	gitOrigin, err := initGitRepo("ciux-git-refresh-test-")
	require.NoError(err)
	rootOrigin, err := gitOrigin.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(rootOrigin)
	_, _, err = gitOrigin.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	depsPath, err := os.MkdirTemp(os.TempDir(), "ciux-git-refresh-deps-")
	require.NoError(err)
	defer os.RemoveAll(depsPath)

	// First run: nothing in the base path yet, the dependency is cloned
	gitObj := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: "master",
	}
	err = gitObj.CloneOrOpen(depsPath, true, false)
	require.NoError(err)
	require.False(gitObj.InPlace)

	// The dependency moves on: work branch created, with a newer tag
	branchName := "testbranch"
	err = gitOrigin.CreateBranch(branchName)
	require.NoError(err)
	_, _, err = gitOrigin.TaggedCommit("second.txt", "second", "v2.0.0", true, author)
	require.NoError(err)

	// Second run in the same base path: the clone is left untouched, so it is
	// still on master and its version is the stale one
	inPlace := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: branchName,
	}
	err = inPlace.CloneOrOpen(depsPath, true, false)
	require.NoError(err)
	require.True(inPlace.InPlace)
	require.Equal("master", inPlace.LocalBranch)
	rev, err := inPlace.GetHeadRevision()
	require.NoError(err)
	require.Equal("v1.0.0", rev.Tag)

	// Same base path, with refresh: cloned again on the work branch
	refreshed := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: branchName,
	}
	err = refreshed.CloneOrOpen(depsPath, true, true)
	require.NoError(err)
	require.False(refreshed.InPlace)
	head, err := refreshed.Repository.Head()
	require.NoError(err)
	require.Equal(branchName, head.Name().Short())
	rev, err = refreshed.GetHeadRevision()
	require.NoError(err)
	require.Equal("v2.0.0", rev.Tag)
}

// Refreshing removes a clone, so it must never sacrifice uncommitted work
func TestCloneOrOpenRefreshDirty(t *testing.T) {
	require := require.New(t)

	gitOrigin, err := initGitRepo("ciux-git-refreshdirty-test-")
	require.NoError(err)
	rootOrigin, err := gitOrigin.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(rootOrigin)
	_, _, err = gitOrigin.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)

	depsPath, err := os.MkdirTemp(os.TempDir(), "ciux-git-refreshdirty-deps-")
	require.NoError(err)
	defer os.RemoveAll(depsPath)

	gitObj := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: "master",
	}
	err = gitObj.CloneOrOpen(depsPath, true, false)
	require.NoError(err)
	cloneRoot, err := gitObj.GetRoot()
	require.NoError(err)

	// Local changes in the clone: refreshing must fail instead of removing it
	err = os.WriteFile(filepath.Join(cloneRoot, "first.txt"), []byte("work in progress"), 0644)
	require.NoError(err)

	dirty := &Git{
		Url:        "file://" + rootOrigin,
		WorkBranch: "master",
	}
	err = dirty.CloneOrOpen(depsPath, true, true)
	require.ErrorContains(err, "local changes")
	_, err = os.Stat(cloneRoot)
	require.NoError(err)
}

func TestMainBranch(t *testing.T) {
	require := require.New(t)

	var mainBranch string
	gitLocal, err := initGitRepo("ciux-git-mainbranch-test-")
	require.NoError(err)
	headHash, _, err := gitLocal.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	mainBranch, hash, err := gitLocal.MainBranch()
	require.NoError(err)
	require.Equal("master", mainBranch)
	require.Equal(headHash.String(), hash)

	root, err := gitLocal.GetRoot()
	require.NoError(err)

	tests := []struct {
		name     string
		url      string
		clone    bool
		expected string
	}{
		{
			name:     "master",
			url:      "file://" + root,
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
			err := gitObj.CloneOrOpen("", false, false)
			require.NoError(err)
		}
		mainBranch, _, err = gitObj.MainBranch()
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
func TestGoInstall(t *testing.T) {
	require := require.New(t)

	git, err := initGitRepo("ciux-git-go-install-test-")
	require.NoError(err)
	root, err := git.GetRoot()
	require.NoError(err)

	cmd := fmt.Sprintf("go mod init -C %s ciux-git-go-install-test/ciux-fake", root)
	_, _, err = ExecCmd(cmd, false)
	require.NoError(err)

	// Create main.go in root
	fname := "main.go"
	f, err := os.Create(filepath.Join(root, fname))
	require.NoError(err)
	_, err = f.WriteString("package main\n\nfunc main() {\n\tprintln(\"Hello world\")\n}\n")
	require.NoError(err)
	f.Close()

	err = git.GoInstall()
	require.NoError(err)

	// Clean up
	cmd = "rm $(which ciux-fake)"
	_, _, err = ExecCmd(cmd, false)
	require.NoError(err)

	// TODO: Add assertions to verify the behavior of the GoInstall function
}
func TestGetRevision(t *testing.T) {
	require := require.New(t)

	// Create a new Git repository
	gitObj, err := prepareTestRepository()
	require.NoError(err)
	root, err := gitObj.GetRoot()
	require.NoError(err)

	hash1, err := gitObj.Repository.ResolveRevision("v1.0.0")
	require.NoError(err)

	hash3, err := gitObj.Repository.ResolveRevision("HEAD")
	require.NoError(err)

	// Call the GetRevision method
	revision, err := gitObj.GetRevision(*hash1)
	require.NoError(err)

	// Verify the returned GitRevision object
	require.Equal("v1.0.0", revision.Tag)
	require.Equal(0, revision.Counter)
	require.Equal(hash1.String(), revision.Hash)
	require.False(revision.Dirty)

	// Call the GetRevision method
	revision, err = gitObj.GetRevision(*hash3)
	require.NoError(err)

	// Verify the returned GitRevision object
	require.Equal("v2.0.0", revision.Tag)
	require.Equal(1, revision.Counter)
	require.Equal(hash3.String(), revision.Hash)
	require.False(revision.Dirty)

	os.RemoveAll(root)
}
func TestGetRoot(t *testing.T) {
	require := require.New(t)

	repoPrefix := "ciux-git-getroot-test-"

	// Test with a valid repository
	gitObj, err := initGitRepo(repoPrefix)
	require.NoError(err)
	root, err := gitObj.GetRoot()
	require.NoError(err)
	require.NotEmpty(root)
	// The returned root should be a directory and should exist
	info, err := os.Stat(root)
	require.NoError(err)
	require.True(info.IsDir())
	// Check root starts with the expected prefix

	require.True(HasPrefixInBase(root, repoPrefix))
}
