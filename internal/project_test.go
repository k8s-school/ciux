package internal

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var finkBrokerRepoUrl = "https://github.com/astrolabsoftware/fink-broker"
var finkBrokerRegistryURL = "gitlab-registry.in2p3.fr/astrolabsoftware/fink"

func setupFinkBrokerProject() (Project, error) {
	// Test with fink-broker repository
	gitObj := &Git{
		Url:        finkBrokerRepoUrl,
		WorkBranch: "master",
	}
	err := gitObj.CloneOrOpen("", true, false)
	if err != nil {
		return Project{}, err
	}
	root, err := gitObj.GetRoot()
	if err != nil {
		return Project{}, err
	}

	project, err := NewProject(root, "", false, "")
	project.ImageRegistry = finkBrokerRegistryURL
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func setupTestProject(pattern string) (Git, []Git, ProjConfig, error) {

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
				Url:    "https://github.com/k8s-school/ktbx",
				Clone:  false,
				Pull:   false,
				Labels: map[string]string{"build": "true"},
			},
			{
				Url:    "https://github.com/k8s-school/k8s-server",
				Clone:  false,
				Pull:   false,
				Labels: map[string]string{"build": "false"},
			},
			{
				Image:  "https://github.com/k8s-school/ktbx",
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
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{Author: &author})
	if err != nil {
		return Git{}, []Git{}, ProjConfig{}, err
	}

	// Create a new tag for the commit
	_, err = gitMeta.Repository.CreateTag("v1.0.0", commit, &git.CreateTagOptions{
		Tagger:  &author,
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

	localGit, remoteGitDeps, _, err := setupTestProject("ciux-scanremotedeps-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(root)
	depRoot, err := remoteGitDeps[0].GetRoot()
	require.NoError(err)
	defer os.RemoveAll(depRoot)

	project, err := NewProject(root, "", false, "")
	require.NoError(err)

	// Assert that the dependency has the correct branch information
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the main repository
	branchName := "testbranch"
	err = localGit.CreateBranch(branchName)
	require.NoError(err)

	err = project.scanRemoteDeps()
	require.NoError(err)
	require.Equal("master", project.Dependencies[0].Git.WorkBranch)

	// Create testbranch in the dependency repository
	err = remoteGitDeps[0].CreateBranch(branchName)
	require.NoError(err)

	err = project.scanRemoteDeps()
	require.NoError(err)
	require.Equal("testbranch", project.Dependencies[0].Git.WorkBranch)

}
func TestWriteOutConfig(t *testing.T) {
	require := require.New(t)

	localGit, _, _, err := setupTestProject("ciux-writeoutconfig-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)

	project, err := NewProject(root, "", false, "build=true")
	require.NoError(err)
	tmpDir, err := os.MkdirTemp("", "ciux-writeoutconfig-test-projectdeps-")
	require.NoError(err)
	project.RetrieveDepsSources(tmpDir, false)
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
		if !git.isRemoteOnly() {
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
	localGit, _, projConfig, err := setupTestProject(patternDir)
	require.NoError(err)
	repoDir, err := localGit.GetRoot()
	require.NoError(err)

	// Create a new project using the test repository and configuration file
	project, err := NewProject(repoDir, "", false, "")
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
	project, err = NewProject(repoDir, "", false, "build=true")
	require.NoError(err)
	require.Len(project.Dependencies, 2)
}

func TestGetImageName(t *testing.T) {
	require := require.New(t)

	registry := "test-registry.io"
	ciRegistry := "ci-internal-registry.io"

	localGit, _, _, err := setupTestProject("ciux-getimage-test-")
	require.NoError(err)
	root, err := localGit.GetRoot()
	require.NoError(err)
	defer os.RemoveAll(root)

	project, err := NewProject(root, "", false, "")
	require.NoError(err)

	// Test when checkRegistry is true and image is not found in the registry
	// with no ci internal repository
	project.ImageRegistry = registry
	err = project.GetImageName("", true)
	require.NoError(err)
	image := project.Image
	require.False(image.InRegistry)
	require.Equal(registry, image.Registry, "Invalid registr for image %s", image)
	require.Containsf(image.String(), "ciux-getimage-test-main-", "Invalid name for image %s", image)
	require.Equal("v1.0.0", image.Tag)
	t.Logf("Image %s:", image)

	// Test when checkRegistry is true and image is notfound in the registry
	// with ci internal repository*
	project.TemporaryRegistry = ciRegistry
	suffix := "noscience"
	err = project.GetImageName(suffix, true)
	require.NoError(err)
	image = project.Image
	require.False(image.InRegistry)
	require.Equal(ciRegistry, image.Registry)

	require.Containsf(image.String(), "ciux-getimage-test-main-", "Invalid name for image %s", image)
	require.Containsf(image.String(), "-noscience", "Invalid name for image %s", image)
	require.Equal("v1.0.0", image.Tag)
	t.Logf("Image %s:", image)

	// Test when checkRegistry is false
	err = project.GetImageName("", false)
	require.NoError(err)
	image = project.Image
	require.False(image.InRegistry)
	require.Equal(ciRegistry, image.Registry)
	require.NotEmpty(image.Name)
	require.Equal("v1.0.0", image.Tag)
}

func TestGetImageNameFinkBroker(t *testing.T) {
	require := require.New(t)

	project, err := setupFinkBrokerProject()
	require.NoError(err)

	w, err := project.GitMain.Repository.Worktree()
	require.NoError(err)

	// Test when image is found in the registry for the current commit
	hash := "655447f544bc4e26ea0c5d23d35ac4772d20ed73"
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(hash),
	})
	require.NoError(err)
	err = project.GetImageName("no-science", false)
	require.NoError(err)
	image := project.Image
	t.Logf("Image %s:", image)
	require.False(image.InRegistry)
	//require.Equal(ciRegistry, image.Registry)
	require.NotEmpty(image.Name)
	require.Equal("v3.1.1-rc1-7-g655447f", image.Tag)
}

// commitFile commits an empty file, without tag
func commitFile(t *testing.T, gitObj Git, filename string) plumbing.Hash {
	worktree, err := gitObj.Repository.Worktree()
	require.NoError(t, err)
	root, err := gitObj.GetRoot()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, filename), nil, 0644))
	_, err = worktree.Add(filename)
	require.NoError(t, err)
	hash, err := worktree.Commit("Add "+filename, &git.CommitOptions{Author: &author})
	require.NoError(t, err)
	return hash
}

func TestFindInRegistryImage(t *testing.T) {
	require := require.New(t)

	server := httptest.NewServer(registry.New())
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")

	gitObj, err := initGitRepo("ciux-findinregistry-test-")
	require.NoError(err)
	hash1, _, err := gitObj.TaggedCommit("first.txt", "first", "v1.0.0", true, author)
	require.NoError(err)
	hash2 := commitFile(t, gitObj, "second.txt")
	hash3 := commitFile(t, gitObj, "third.txt")
	hashes := []plumbing.Hash{*hash1, hash2, hash3}

	project := Project{GitMain: &gitObj, ImageRegistry: host}

	// Only the image of the second commit is in the registry
	rev, err := gitObj.GetRevision(hash2)
	require.NoError(err)
	tag := rev.GetVersion()
	require.Equal("v1.0.0-1-g"+hash2.String()[:7], tag)
	base := randomImage(t)
	ref, err := name.ParseReference(fmt.Sprintf("%s/myimage:%s", host, tag))
	require.NoError(err)
	require.NoError(remote.Write(ref, builtOn(t, base)))
	urlBase := pushImage(t, host, "base", base)
	urlOtherBase := pushImage(t, host, "other-base", randomImage(t))

	tests := []struct {
		name      string
		imageName string
		baseImage string
		found     bool
	}{
		{"image of one of the commits, no base declared", "myimage", "", true},
		{"image built on the base", "myimage", urlBase, true},
		{"image not built on the base", "myimage", urlOtherBase, false},
		{"image not in the registry", "non-existent-image", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project.BaseImage = tt.baseImage
			image, err := project.findInRegistryImage(tt.imageName, hashes)
			require.NoError(err)
			if !tt.found {
				require.Nil(image)
				return
			}
			require.NotNil(image)
			require.True(image.InRegistry)
			require.Equal(host, image.Registry)
			require.Equal(tt.imageName, image.Name)
			require.Equal(tag, image.Tag)
		})
	}
}

func TestProjectGetName(t *testing.T) {

	tests := []struct {
		name           string
		projectConfig  string
		expectedName   string
		useConfigName  bool
	}{
		{
			name:          "project name from config",
			projectConfig: "project: my-custom-project\nregistry: test-registry",
			expectedName:  "my-custom-project",
			useConfigName: true,
		},
		{
			name:          "fallback to directory name when no project in config",
			projectConfig: "registry: test-registry",
			expectedName:  "", // Will be set to directory name in test
			useConfigName: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "ciux-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Initialize git repo
			_, err = git.PlainInit(tmpDir, false)
			require.NoError(t, err)

			// Create .ciux config file
			configPath := filepath.Join(tmpDir, ".ciux")
			err = os.WriteFile(configPath, []byte(tt.projectConfig), 0644)
			require.NoError(t, err)

			// Create a project
			project, _, err := NewCoreProject(tmpDir, "")
			require.NoError(t, err)

			// Test GetName
			actualName, err := project.GetName()
			require.NoError(t, err)

			if tt.useConfigName {
				require.Equal(t, tt.expectedName, actualName)
			} else {
				// When no project name in config, should fallback to directory name
				require.Equal(t, filepath.Base(tmpDir), actualName)
			}
		})
	}
}

// The generated shell configuration file must tell that a dependency was not
// cloned by this run, and warn when its version is then not the one of the
// work branch
func TestInPlaceComment(t *testing.T) {
	require := require.New(t)

	cloned := &Git{WorkBranch: "master"}
	require.Empty(inPlaceComment(cloned, "FINK_ALERT_SIMULATOR"))

	onWorkBranch := &Git{InPlace: true, WorkBranch: "master", LocalBranch: "master"}
	comment := inPlaceComment(onWorkBranch, "FINK_ALERT_SIMULATOR")
	require.Contains(comment, "# FINK_ALERT_SIMULATOR: repository already in the workspace")
	require.NotContains(comment, "WARNING")

	stale := &Git{InPlace: true, WorkBranch: "1216-run-fink-broker", LocalBranch: "master"}
	comment = inPlaceComment(stale, "FINK_ALERT_SIMULATOR")
	require.Contains(comment, "# WARNING: FINK_ALERT_SIMULATOR_VERSION below comes from branch 'master', "+
		"not from FINK_ALERT_SIMULATOR_WORKBRANCH")

	detached := &Git{InPlace: true, WorkBranch: "master"}
	comment = inPlaceComment(detached, "FINK_ALERT_SIMULATOR")
	require.Contains(comment, "comes from a detached HEAD")
}
