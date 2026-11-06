/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// var pathes []string
var check bool
var env bool
var build string
var tmpRegistry string

// imageCmd represents the revision command
var imageCmd = &cobra.Command{
	Use:     "image (REPOSITORY)",
	Aliases: []string{"img"},
	Short:   "Retrieve the version of a container image, based on the source code used to build it",
	Long: `Retrieve the version of a container image, based on the source code used to build it.
If source code has not been modified in the current commit, ciux will return an previously built image with the current code if this image is available in the registry.
- Use "sourcePathes" in the .ciux configuration file to specify the pathes to source code used to build the container image
this pathes are relatives and must be used in the image's Dockerfile COPY/ADD commands
- Use "registry" in the .ciux configuration file to specify the registry where the image is stored`,
	Example: `# Check if image registry/<project_name>-<build>:<tag> exists and is built on
# the base image of the build, i.e. the dependency with "baseOf: <build>" in .ciux
# tag is in the format vX.Y.Z[-rcT]-N-g<short-commit-hash>
# the image of "defaultBuild" in .ciux has no suffix: registry/<project_name>:<tag>
ciux get image --check <path_to_git_repository> --build <build>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := internal.AbsPath(args[0])
		project, _, err := internal.NewCoreProject(repositoryPath, branch)
		internal.FailOnError(err)
		project.TemporaryRegistry = tmpRegistry
		err = project.GetImageName(build, check)
		internal.FailOnError(err)

		if env {
			fmt.Printf("export CIUX_IMAGE_URL=%s\n", project.Image.Url())
			fmt.Printf("export CIUX_BUILD=%t\n", !project.Image.InRegistry)
			fmt.Printf("export CIUX_BASE_IMAGE=%s\n", project.BaseImage)
		} else {
			fmt.Printf("Image: %s\n", project.Image)
		}
	},
}

func init() {
	getCmd.AddCommand(imageCmd)

	//imageTagCmd.Flags().StringSliceVarP(&pathes, "pathes", "p", []string{"rootfs"}, "Relative pathes to source code used to build the container image")
	imageCmd.Flags().BoolVarP(&check, "check", "c", false, "Check if an image with same source code is already available in the registry, if not exit with error and print the name of the image to build")
	addBuildFlags(imageCmd)
	imageCmd.Flags().StringVarP(&tmpRegistry, "tmp-registry", "t", "", "Name of temporary registry used to store the image during the ci process")
	imageCmd.Flags().BoolVarP(&env, "env", "e", false, "Print environment variables to use the image in the CI process, CIUX_IMAGE_URL, CIUX_BUILD and CIUX_BASE_IMAGE")
}

// addBuildFlags adds the --build flag, and --suffix, its deprecated alias
func addBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&build, "build", "",
		"Build of the image: it is the suffix of the image name, except for \"defaultBuild\" in .ciux, "+
			"and selects the base image, the dependency with \"baseOf: <build>\" in .ciux")
	cmd.Flags().StringVarP(&build, "suffix", "p", "", "Suffix to add to the image name")
	internal.FailOnError(cmd.Flags().MarkDeprecated("suffix", "use --build"))
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
