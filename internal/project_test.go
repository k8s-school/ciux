package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestGetDepsBranches(t *testing.T) {
	assert := assert.New(t)

	// Create a temporary directory for the project git repository
	gitMeta, err := initGitRepo("ciux-git-getversions-test-")
	assert.NoError(err)
	root, err := gitMeta.GetRoot()
	assert.NoError(err)
	defer os.RemoveAll(root)

	// Create a custom .ciux file in the repository
	pattern := "ciux-git-getversions-test-"
	gitDepMeta, err := initGitRepo(pattern)
	assert.NoError(err)
	depRoot, err := gitDepMeta.GetRoot()
	defer os.RemoveAll(depRoot)
	assert.NoError(err)

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
	assert.NoError(err)
	newviper := viper.New()
	err = newviper.MergeConfigMap(items)
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

	gitDeps, err := project.GetDepsWorkBranch()
	assert.NoError(err)

	// Assert that the dependency has the correct branch information
	assert.Equal("master", gitDeps[0].WorkBranch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = gitMeta.CreateBranch(branchName)
	assert.NoError(err)

	gitDeps, err = project.GetDepsWorkBranch()
	assert.NoError(err)
	assert.Equal("master", gitDeps[0].WorkBranch)

	// Create testbranch in the dependency repository
	err = gitDepMeta.CreateBranch(branchName)
	assert.NoError(err)

	gitDeps, err = project.GetDepsWorkBranch()
	assert.NoError(err)
	assert.Equal("testbranch", gitDeps[0].WorkBranch)

}
