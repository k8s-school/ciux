/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// igniteCmd represents the revision command
var igniteCmd = &cobra.Command{
	Use:     "ignite repository_path",
	Aliases: []string{"ig", "ign"},
	Short:   "Prepare integration test",
	Long: `Retrieve current revision of the repository and clone all dependencies in the correct revision.
	Check if dependencies container images are available.
	Use repository_path/.ciux.yaml configuration file to retrieve dependencies.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		project := internal.NewProject(repositoryPath)
		msg, err := internal.String(repositoryPath)
		internal.FailOnError(err)
		internal.Info(msg)

		gitDeps, err := project.GetDepsWorkBranch()
		internal.FailOnError(err)
		for _, gitDep := range gitDeps {
			singleBranch := true
			gitDep.Clone(singleBranch)
			rev, err := gitDep.GetRevision()
			internal.FailOnError(err)
			internal.Info("Dep repo: %s, version: %+v", gitDep.Url, rev.GetVersion())
		}
	},
}

func init() {
	rootCmd.AddCommand(igniteCmd)
}
