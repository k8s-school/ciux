/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// revisionCmd represents the revision command
var revisionCmd = &cobra.Command{
	Use:   "revision (REPOSITORY) (DEPENDENCY_REPOSITORIES...)",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		gitMeta, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)
		rev, err := gitMeta.GetRevision()
		internal.FailOnError(err)
		internal.Infof("Revision: %+v", rev)
	},
}

func init() {
	getCmd.AddCommand(revisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
