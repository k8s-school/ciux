package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/k8s-school/ciux/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func createTestProject(pattern string) (Project, error) {
	// Create a temporary directory for the project git repository
	gitMeta, err := initGitRepo(pattern + "main-")
	if err != nil {
		return Project{}, err
	}
	root, err := gitMeta.GetRoot()
	if err != nil {
		return Project{}, err
	}

	// Create a custom .ciux file in the repository
	gitDepMeta, err := initGitRepo(pattern + "dep-")
	if err != nil {
		return Project{}, err
	}
	depRoot, err := gitDepMeta.GetRoot()
	defer os.RemoveAll(depRoot)
	if err != nil {
		return Project{}, err
	}

	// Write .ciux file in project directory
	config := Config{
		Registry: "test-registry.io",
		Dependencies: []Dependency{
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

	project := Project{
		Git:    gitMeta,
		Config: config,
	}

	items := map[string]interface{}{}
	err = mapstructure.Decode(config, &items)
	if err != nil {
		return project, err
	}
	newviper := viper.New()
	err = newviper.MergeConfigMap(items)
	if err != nil {
		return project, err
	}

	yamlData, err := yaml.Marshal(items)
	if err != nil {
		return project, err
	}
	log.Debugf("yamlData: %s", string(yamlData))
	ciuxPath := filepath.Join(root, ".ciux")
	f, err := os.Create(ciuxPath)
	if err != nil {
		return project, err
	}
	_, err = f.Write(yamlData)
	if err != nil {
		return project, err
	}
	f.Close()

	// Add the file to the repository
	worktree, err := gitMeta.Repository.Worktree()
	if err != nil {
		return project, err
	}
	_, err = worktree.Add(".ciux")
	if err != nil {
		return project, err
	}

	// Commit the changes to the repository
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		return project, err
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
		return project, err
	}

	// Initialize the dependency repository
	_, _, err = gitDepMeta.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	if err != nil {
		return project, err
	}
	return project, nil
}

func TestGetDepsBranches(t *testing.T) {
	assert := assert.New(t)

	project, err := createTestProject("ciux-getdepsbranches-test-")
	root, err := project.Git.GetRoot()
	assert.NoError(err)
	defer os.RemoveAll(root)
	depRoot, err := project.GitDeps[0].GetRoot()
	assert.NoError(err)
	defer os.RemoveAll(depRoot)

	err = project.GetDepsWorkBranch()
	assert.NoError(err)

	// Assert that the dependency has the correct branch information
	assert.Equal("master", project.GitDeps[0].WorkBranch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = project.Git.CreateBranch(branchName)
	assert.NoError(err)

	err = project.GetDepsWorkBranch()
	assert.NoError(err)
	assert.Equal("master", project.GitDeps[0].WorkBranch)

	// Create testbranch in the dependency repository
	err = project.GitDeps[0].CreateBranch(branchName)
	assert.NoError(err)

	err = project.GetDepsWorkBranch()
	assert.NoError(err)
	assert.Equal("testbranch", project.GitDeps[0].WorkBranch)

}
func TestWriteOutConfig(t *testing.T) {
	assert := assert.New(t)

	project, err := createTestProject("ciux-writeoutconfig-test-")
	root, err := project.Git.GetRoot()
	assert.NoError(err)
	defer os.RemoveAll(root)
	depRoot, err := project.GitDeps[0].GetRoot()
	assert.NoError(err)
	defer os.RemoveAll(depRoot)

	err = project.GetDepsWorkBranch()
	assert.NoError(err)

	err = project.WriteOutConfig()
	assert.NoError(err)

	// Assert that the .ciux.sh file was created
	ciuxConfig := filepath.Join(root, "ciux.sh")
	_, err = os.Stat(ciuxConfig)
	assert.NoError(err)

	// Assert that the file contains the expected environment variables
	expectedVars := []string{
		"TEST-REPO=" + root,
		"TEST-REPO-DEPENDENCIES=" + depRoot,
	}
	f, err := os.Open(ciuxConfig)
	assert.NoError(err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		notFoundVars := expectedVars
		for i, expectedVar := range expectedVars {
			if strings.Contains(line, expectedVar) {
				notFoundVars = slices.Delete(notFoundVars, i, i)
			}
		}
		expectedVars = notFoundVars
	}
	assert.Empty(expectedVars)
}
