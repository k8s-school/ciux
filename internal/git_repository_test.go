package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

var rootfs = "rootfs"

func prepareTestRepository() (Git, error) {
	// Create a new Git repository
	gitObj, err := initGitRepo("ciux-hasdiff-test-")
	if err != nil {
		return Git{}, err
	}

	root, err := gitObj.GetRoot()
	if err != nil {
		return Git{}, err
	}

	// Create some commits with code changes

	err = os.MkdirAll(filepath.Join(root, rootfs), 0755)
	if err != nil {
		return Git{}, err
	}

	filename1 := filepath.Join(rootfs, "file1.txt")
	_, _, err = gitObj.TaggedCommit(filename1, "commit1", "v1.0.0", true, author)
	if err != nil {
		return Git{}, err
	}

	_, _, err = gitObj.TaggedCommit("file2.txt", "commit2", "v2.0.0", true, author)
	if err != nil {
		return Git{}, err
	}

	worktree, err := gitObj.Repository.Worktree()
	if err != nil {
		return Git{}, err
	}

	d1 := []byte("hello\ngo\n")
	err = os.WriteFile(filepath.Join(root, filename1), d1, 0644)
	if err != nil {
		return Git{}, err
	}

	_, err = worktree.Add(filename1)
	if err != nil {
		return Git{}, err
	}

	_, err = worktree.Commit("update file1", &git.CommitOptions{Author: &author})
	if err != nil {
		return Git{}, err
	}

	return gitObj, nil
}

func TestHasDiff(t *testing.T) {
	require := require.New(t)

	gitObj, err := prepareTestRepository()
	require.NoError(err)
	repo := gitObj.Repository

	hash1, err := gitObj.Repository.ResolveRevision("v1.0.0")
	require.NoError(err)

	hash2, err := gitObj.Repository.ResolveRevision("v2.0.0")
	require.NoError(err)

	hash3, err := gitObj.Repository.ResolveRevision("HEAD")
	require.NoError(err)

	// Test case 1: File not changed in the commit
	currentCommit, err := repo.CommitObject(*hash2)
	require.NoError(err)
	ancestorCommit, err := repo.CommitObject(*hash1)
	require.NoError(err)
	pathes := []string{rootfs}
	hasDiff, err := HasDiff(currentCommit, ancestorCommit, pathes)
	require.NoError(err)
	require.False(hasDiff)

	pathes = []string{""}
	hasDiff, err = HasDiff(currentCommit, ancestorCommit, pathes)
	require.NoError(err)
	require.True(hasDiff)

	// Test case 1: File not changed in the commit
	currentCommit, err = repo.CommitObject(*hash3)
	require.NoError(err)
	ancestorCommit, err = repo.CommitObject(*hash1)
	require.NoError(err)
	pathes = []string{rootfs}
	hasDiff, err = HasDiff(currentCommit, ancestorCommit, pathes)
	require.NoError(err)
	require.True(hasDiff)
}

func TestFindCodeChange(t *testing.T) {
	require := require.New(t)

	// Create a new Git repository
	gitObj, err := initGitRepo("ciux-latestcommit-test-")
	require.NoError(err)
	repo := gitObj.Repository

	// Create some commits with code changes
	root, err := gitObj.GetRoot()
	require.NoError(err)
	rootfs := "rootfs"
	err = os.MkdirAll(filepath.Join(root, rootfs), 0755)
	require.NoError(err)
	hash1, _, err := gitObj.TaggedCommit("rootfs/file1.txt", "commit1", "v1.0.0", true, author)
	require.NoError(err)

	hash2, _, err := gitObj.TaggedCommit("rootfs/file2.txt", "commit2", "v2.0.0", true, author)
	require.NoError(err)

	hash3, _, err := gitObj.TaggedCommit("file3.txt", "commit3", "v3.0.0", true, author)
	require.NoError(err)

	hash4, _, err := gitObj.TaggedCommit("file4.txt", "commit4", "v4.0.0", true, author)
	require.NoError(err)

	// Test case 1: latest commit with code change is the latest commit
	firstChangeCommitHash, inBetweenCommits, err := FindCodeChange(repo, *hash1, []string{"rootfs"})
	require.NoError(err)
	require.Equal(*hash1, firstChangeCommitHash)
	require.Len(inBetweenCommits, 0)

	// Test case 2: latest commit with code change is not the latest commit
	firstChangeCommitHash, inBetweenCommits, err = FindCodeChange(repo, *hash4, []string{"rootfs"})
	require.NoError(err)
	require.Equal(*hash2, firstChangeCommitHash)
	require.Len(inBetweenCommits, 2)
	require.Contains(inBetweenCommits, *hash3)
	require.Contains(inBetweenCommits, *hash4)
}
