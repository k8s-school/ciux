/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var pathes []string
var full bool

// imageTagCmd represents the revision command
var imageTagCmd = &cobra.Command{
	Use:     "imagetag (REPOSITORY)",
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

		if full {
			repositoryPath := args[0]
			project, err = internal.NewProject(repositoryPath, branch, labelSelector)
			internal.FailOnError(err)
			gitMain = project.GitMain
		} else {
			gitMain, err = internal.NewGit(repositoryPath)
			internal.FailOnError(err)
		}
		head, err := gitMain.Repository.Head()
		internal.FailOnError(err)
		commit, err := internal.FindCodeChange(gitMain.Repository, head.Hash(), pathes)
		internal.FailOnError(err)
		rev, err := gitMain.GetRevision(commit.Hash)
		internal.FailOnError(err)
		if full {
			name, err := gitMain.GetName()
			internal.FailOnError(err)
			imageUrl := fmt.Sprintf("%s/%s:%s", project.ImageRegistry, name, rev.GetVersion())
			fmt.Printf("%s\n", imageUrl)
		} else {
			fmt.Printf("%s\n", rev.GetVersion())
		}
	},
}

func init() {
	getCmd.AddCommand(imageTagCmd)

	imageTagCmd.Flags().StringSliceVarP(&pathes, "pathes", "p", []string{"rootfs"}, "Relative pathes to source code used to build the container image")
	imageTagCmd.Flags().BoolVarP(&full, "full", "f", false, "Display full image name")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
