/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		repositoryPath := internal.AbsPath(args[0])
		git, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)

		rev, err := git.GetHeadRevision()
		internal.FailOnError(err)
		newTag, err := rev.UpgradeTag()
		internal.FailOnError(err)
		msg := fmt.Sprintf("git tag -m \"Release %[1]s\" %[1]s\n", newTag)
		msg += "git push --tag"
		internal.Infof(msg)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
