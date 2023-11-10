/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

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

		depsDir := filepath.Dir(repositoryPath)

		// Clone dependencies directories and checkout the correct revision
		// Check container images exist
		err := project.ScanRemoteDeps()
		internal.FailOnError(err)
		err = project.SetDepsRepos(depsDir)
		internal.FailOnError(err)
		internal.Infof("%s", project.String())
		images, err := project.CheckImages()
		internal.FailOnError(err)
		internal.Infof("Images: %v", images)
		err = project.WriteOutConfig()
		internal.FailOnError(err)
	},
}

func init() {
	rootCmd.AddCommand(igniteCmd)
}
