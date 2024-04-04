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
var suffix string
var tmpRegistry string

// imageCmd represents the revision command
var imageCmd = &cobra.Command{
	Use:     "image (REPOSITORY)",
	Aliases: []string{"img"},
	Short:   "Retrieve the version of a container image, based on the source code used to build it",
	Long: `Retrieve the version of a container image, based on the source code used to build it
	  Use --pathes to specify the pathes to source code used to build the container image
	  this pathes are relatives and must be used in the image's Dockerfile COPY/ADD commands`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		project, _, err := internal.NewCoreProject(repositoryPath, branch)
		project.TemporaryRegistry = tmpRegistry
		internal.FailOnError(err)
		image, err := project.GetImageName(suffix, check)
		internal.FailOnError(err)

		if env {
			fmt.Printf("export CIUX_IMAGE_URL=%s\n", image.Url())
			fmt.Printf("export CIUX_BUILD=%t\n", !image.InRegistry)
		} else {
			fmt.Printf("Image: %s\n", image)
		}
	},
}

func init() {
	getCmd.AddCommand(imageCmd)

	//imageTagCmd.Flags().StringSliceVarP(&pathes, "pathes", "p", []string{"rootfs"}, "Relative pathes to source code used to build the container image")
	imageCmd.Flags().BoolVarP(&check, "check", "c", false, "Check if an image with same source code is already available in the registry, if not exit with error and print the name of the image to build")
	imageCmd.Flags().StringVarP(&suffix, "suffix", "p", "", "Suffix to add to the image name")
	imageCmd.Flags().StringVarP(&tmpRegistry, "tmp-registry", "t", "", "Name of temporary registry used to store the image during the ci process")
	imageCmd.Flags().BoolVarP(&env, "env", "e", false, "Print environment variables to use the image in the CI process, CIUX_IMAGE_URL and CIUX_BUILD")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
