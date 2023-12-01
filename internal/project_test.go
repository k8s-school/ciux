package internal

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func prepareTestProject(pattern string) (Git, []Git, error) {

	// Create a temporary directory for the project git repository
	gitMeta, err := initGitRepo(pattern + "main-")
	if err != nil {
		return Git{}, []Git{}, err
	}
	root, err := gitMeta.GetRoot()
	if err != nil {
		return Git{}, []Git{}, err
	}

	// Create a custom .ciux file in the repository
	gitDepMeta, err := initGitRepo(pattern + "dep-")
	if err != nil {
		return Git{}, []Git{}, err
	}
	depRoot, err := gitDepMeta.GetRoot()
	if err != nil {
		return Git{}, []Git{}, err
	}

	// Write .ciux file in project directory
	config := ProjConfig{
		Registry: "test-registry.io",
		Dependencies: []DepConfig{
			{
				Url:   "file://" + depRoot,
				Clone: true,
				Pull:  true,
			},
			{
				Url:   "https://github.com/k8s-school/k8s-toolbox",
				Clone: false,
				Pull:  false,
			},
		},
	}

	items := map[string]interface{}{}
	err = mapstructure.Decode(config, &items)
	if err != nil {
		return Git{}, []Git{}, err
	}
	newviper := viper.New()
	err = newviper.MergeConfigMap(items)
	if err != nil {
		return Git{}, []Git{}, err
	}

	yamlData, err := yaml.Marshal(items)
	if err != nil {
		return Git{}, []Git{}, err
	}
	slog.Debug("Ciux configuration", "yaml", string(yamlData))
	ciuxPath := filepath.Join(root, ".ciux")
	f, err := os.Create(ciuxPath)
	if err != nil {
		return Git{}, []Git{}, err
	}
	_, err = f.Write(yamlData)
	if err != nil {
		return Git{}, []Git{}, err
	}
	f.Close()

	// Add the file to the repository
	worktree, err := gitMeta.Repository.Worktree()
	if err != nil {
		return Git{}, []Git{}, err
	}
	_, err = worktree.Add(".ciux")
	if err != nil {
		return Git{}, []Git{}, err
	}

	// Commit the changes to the repository
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		return Git{}, []Git{}, err
	}

	// Create a new tag for the commit
	_, err = gitMeta.Repository.CreateTag("v1.0.0", commit, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Message: "Test tag",
	})
	if err != nil {
		return Git{}, []Git{}, err
	}

	// Initialize the dependency repository
	_, _, err = gitDepMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	if err != nil {
		return Git{}, []Git{}, err
	}
	return gitMeta, []Git{gitDepMeta}, nil
}

func TestScanRemoteDeps(t *testing.T) {
	require := require.New(t)

	localGit, remoteGitDeps, err := prepareTestProject("ciux-scanremotedeps-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(root)
	depRoot, err := remoteGitDeps[0].GetRoot()
	require.NoError(err)
	defer os.RemoveAll(depRoot)

	project := NewProject(root)

	// Assert that the dependency has the correct branch information
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = localGit.CreateBranch(branchName)
	require.NoError(err)

	err = project.ScanRemoteDeps()
	require.NoError(err)
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the dependency repository
	err = remoteGitDeps[0].CreateBranch(branchName)
	require.NoError(err)

	err = project.ScanRemoteDeps()
	require.NoError(err)
	require.Equal("testbranch", project.Dependencies[0].Git.WorkBranch)

}
func TestWriteOutConfig(t *testing.T) {
	require := require.New(t)

	localGit, _, err := prepareTestProject("ciux-writeoutconfig-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)

	project := NewProject(root)
	tmpDir, err := os.MkdirTemp("", "ciux-writeoutconfig-test-projectdeps-")
	require.NoError(err)
	project.RetrieveDepsSources(tmpDir)
	ciuxConfig := filepath.Join(root, "ciux.sh")
	os.Setenv("CIUXCONFIG", ciuxConfig)
	_, err = project.WriteOutConfig()
	require.NoError(err)

	// Assert that the .ciux.sh file was created
	t.Logf("ciuxConfig: %s", ciuxConfig)
	_, err = os.Stat(ciuxConfig)
	require.NoError(err)

	// Assert that the file contains the expected environment variables
	expectedVars := []string{}
	for _, git := range project.GetGits() {
		if !git.isRemote() {
			varName, err := git.GetEnVarPrefix()
			require.NoError(err)
			depRoot, err := git.GetRoot()
			require.NoError(err)
			expectedVars = []string{
				"export " + varName + "_DIR=" + depRoot,
				"export " + varName + "_VERSION=v1.0.0",
			}
			defer os.RemoveAll(depRoot)
		}
	}
	f, err := os.Open(ciuxConfig)
	require.NoError(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		notFoundVars := expectedVars
		for i, expectedVar := range expectedVars {
			if line == expectedVar {
				notFoundVars = slices.Delete(notFoundVars, i, i+1)
			}
		}
		expectedVars = notFoundVars
	}
	require.Empty(expectedVars)

	os.RemoveAll(root)

}
