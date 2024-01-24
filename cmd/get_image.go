/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// var pathes []string
var check bool
var suffix string

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
		var project internal.Project
		var err error
		var gitMain *internal.Git
		project, err = internal.NewProject(repositoryPath, branch, labelSelector)
		internal.FailOnError(err)
		gitMain = project.GitMain

		slog.Debug("Project source directories", "sourcePathes", project.SourcePathes)

		head, err := gitMain.Repository.Head()
		internal.FailOnError(err)
		commit, err := internal.FindCodeChange(gitMain.Repository, head.Hash(), project.SourcePathes)
		internal.FailOnError(err)
		rev, err := gitMain.GetRevision(commit.Hash)
		internal.FailOnError(err)

		name, err := gitMain.GetName()
		if len(suffix) > 0 {
			name = fmt.Sprintf("%s-%s", name, suffix)
		}
		image := internal.Image{
			Registry: project.ImageRegistry,
			Name:     name,
			Tag:      rev.GetVersion(),
		}

		internal.FailOnError(err)
		var errcheck error
		if check {
			_, _, errcheck = image.Desc()

		}
		if errcheck != nil {
			slog.Debug("Image not found in registry", "image", image)
			rev, err1 := gitMain.GetHeadRevision()
			internal.FailOnError(err1)
			image.Tag = rev.GetVersion()
			fmt.Println(image)
			os.Exit(1)
		}
		fmt.Println(image)
	},
}

func init() {
	getCmd.AddCommand(imageCmd)

	//imageTagCmd.Flags().StringSliceVarP(&pathes, "pathes", "p", []string{"rootfs"}, "Relative pathes to source code used to build the container image")
	imageCmd.Flags().BoolVarP(&check, "check", "c", false, "Check if an image with same source code is already available in the registry, if not exit with error and print the name of the image to build")
	imageCmd.Flags().StringVarP(&suffix, "suffix", "p", "", "Suffix to add to the image name")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
