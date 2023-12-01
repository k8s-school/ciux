/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// releaseCmd represents the revision command
var tagCmd = &cobra.Command{
	Use:     "tag repository_path",
	Aliases: []string{"rel"},
	Short:   "Create a versioned tag for a git repository",
	Long:    `Create a versioned tag for a git repository`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		git, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)

		rev, err := git.GetRevision()
		internal.FailOnError(err)
		newTag, err := rev.UpgradeTag()
		internal.FailOnError(err)
		internal.Infof("New tag: %s", newTag)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
