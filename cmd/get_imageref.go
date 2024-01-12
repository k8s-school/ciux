/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var pathes []string

// imageRefCmd represents the revision command
var imageRefCmd = &cobra.Command{
	Use:     "imageversion (REPOSITORY)",
	Aliases: []string{"img"},
	Short:   "Retrieve the version of a container image, based on the source code used to build it",
	Long: `Retrieve the version of a container image, based on the source code used to build it
	  Use --pathes to specify the pathes to source code used to build the container image
	  this pathes are relatives and must be used in the image's Dockerfile COPY/ADD commands`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		gitMeta, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)
		head, err := gitMeta.Repository.Head()
		internal.FailOnError(err)
		commit, err := internal.FindCodeChange(gitMeta.Repository, head.Hash(), pathes)
		internal.FailOnError(err)
		rev, err := gitMeta.GetRevision(commit.Hash)
		internal.FailOnError(err)
		internal.Infof("TODO: %+v", rev.GetVersion())
	},
}

func init() {
	getCmd.AddCommand(imageRefCmd)

	imageRefCmd.Flags().StringSliceVarP(&pathes, "pathes", "p", []string{"rootfs"}, "Relative pathes to source code used to build the container image")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
