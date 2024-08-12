/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var isrelease bool

// revisionCmd represents the revision command
var revisionCmd = &cobra.Command{
	Use:     "revision (REPOSITORY) (DEPENDENCY_REPOSITORIES...)",
	Aliases: []string{"rev"},
	Short:   "A brief description of your command",
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
		rev, err := gitMeta.GetHeadRevision()
		internal.FailOnError(err)

		if isrelease {
			if rev.IsRelease() {
				internal.Infof(rev.Tag)
			}
		} else {
			internal.Infof("Revision: %+v", rev)
		}
	},
}

func init() {
	getCmd.AddCommand(revisionCmd)

	revisionCmd.Flags().BoolVarP(&isrelease, "isrelease", "r", false, "Check if the current commit is tagged with a release tag and is in master/main branch, return release tag if true, else empty")

}
