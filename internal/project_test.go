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

func prepareTestProject(pattern string) (Git, []Git, ProjConfig, error) {

	// Create a temporary directory for the project git repository
	gitMeta, err := initGitRepo(pattern + "main-")
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	root, err := gitMeta.GetRoot()
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}

	// Create a custom .ciux file in the repository
	gitDepMeta, err := initGitRepo(pattern + "dep-")
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	depRoot, err := gitDepMeta.GetRoot()
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}

	// Write .ciux file in project directory
	config := ProjConfig{
		Registry: "test-registry.io",
		Dependencies: []DepConfig{
			{
				Url:    "file://" + depRoot,
				Clone:  true,
				Pull:   true,
				Labels: map[string]string{"build": "true"},
			},
			{
				Url:    "https://github.com/k8s-school/k8s-toolbox",
				Clone:  false,
				Pull:   false,
				Labels: map[string]string{"build": "true"},
			},
			{
				Url:    "https://github.com/fake/fake-project",
				Clone:  false,
				Pull:   false,
				Labels: map[string]string{"build": "false"},
			},
			{
				Image:  "https://github.com/k8s-school/k8s-toolbox",
				Labels: map[string]string{"ci": "true", "key2": "value2"},
			},
		},
	}

	items := map[string]interface{}{}
	err = mapstructure.Decode(config, &items)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	newviper := viper.New()
	err = newviper.MergeConfigMap(items)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}

	yamlData, err := yaml.Marshal(items)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	slog.Debug("Ciux configuration", "yaml", string(yamlData))
	ciuxPath := filepath.Join(root, ".ciux")
	f, err := os.Create(ciuxPath)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	_, err = f.Write(yamlData)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	f.Close()

	// Add the file to the repository
	worktree, err := gitMeta.Repository.Worktree()
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	_, err = worktree.Add(".ciux")
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}

	// Commit the changes to the repository
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
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
		return Git{}, []Git{}, ProjConfig{}, err
	}

	// Initialize the dependency repository
	_, _, err = gitDepMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}
	return gitMeta, []Git{gitDepMeta}, config, nil
}

func TestScanRemoteDeps(t *testing.T) {
	require := require.New(t)

	localGit, remoteGitDeps, _, err := prepareTestProject("ciux-scanremotedeps-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(root)
	depRoot, err := remoteGitDeps[0].GetRoot()
	require.NoError(err)
	defer os.RemoveAll(depRoot)

	project, err := NewProject(root, "", "")
	require.NoError(err)

	// Assert that the dependency has the correct branch information
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = localGit.CreateBranch(branchName)
	require.NoError(err)

	err = project.ScanRemoteDeps("")
	require.NoError(err)
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the dependency repository
	err = remoteGitDeps[0].CreateBranch(branchName)
	require.NoError(err)

	err = project.ScanRemoteDeps("")
	require.NoError(err)
	require.Equal("testbranch", project.Dependencies[0].Git.WorkBranch)

}
func TestWriteOutConfig(t *testing.T) {
	require := require.New(t)

	localGit, _, _, err := prepareTestProject("ciux-writeoutconfig-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)

	project, err := NewProject(root, "", "build=true")
	require.NoError(err)
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
func TestNewProject(t *testing.T) {
	require := require.New(t)

	patternDir := "ciux-newproject-test-"
	localGit, _, projConfig, err := prepareTestProject(patternDir)
	require.NoError(err)
	repoDir, err := localGit.GetRoot()
	require.NoError(err)

	// Create a new project using the test repository and configuration file
	project, err := NewProject(repoDir, "", "")
	require.NoError(err)

	// Assert the project properties
	root, err := project.GitMain.GetRoot()
	require.NoError(err)
	require.Equal(repoDir, root)
	require.Equal("test-registry.io", project.ImageRegistry)
	require.Len(project.Dependencies, len(projConfig.Dependencies))

	// Assert the first dependency
	require.Contains(project.Dependencies[0].String(), patternDir)

	// Assert the second dependency
	require.Equal(projConfig.Dependencies[1].Url, project.Dependencies[1].String())

	// Assert the third dependency
	require.Equal(projConfig.Dependencies[2].Url, project.Dependencies[2].String())

	// Assert the branch information of the main repository
	branch, err := project.GitMain.GetBranch()
	require.NoError(err)
	require.Equal("master", branch)

	// Create a second project with a selector
	project, err = NewProject(repoDir, "", "build=true")
	require.NoError(err)
	require.Len(project.Dependencies, 2)
}
