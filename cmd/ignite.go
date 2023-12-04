/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var branch string
var itest bool

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
		project := internal.NewProject(repositoryPath, branch)
		depsBasePath := filepath.Dir(repositoryPath)
		// Clone dependencies directories and checkout the correct revision
		// Check container images exist
		if itest {
			err := project.RetrieveDepsSources(depsBasePath)
			internal.FailOnError(err)
			goMsg, err := project.InstallGoModules()
			internal.FailOnError(err)
			internal.Infof("%s", project.String())
			internal.Infof("Go modules installed:\n%s", goMsg)
			images, err := project.CheckImages()
			internal.FailOnError(err)
			internal.Infof("Images: %v", images)
		}
		msg, err := project.WriteOutConfig()
		internal.FailOnError(err)
		internal.Infof("%s", msg)
	},
}

func init() {
	rootCmd.AddCommand(igniteCmd)

	// Here you will define your flags and configuration settings.
	igniteCmd.PersistentFlags().BoolVarP(&itest, "itest", "i", false, "install dependencies for runnning integration tests")
	igniteCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "branch for the project, retrieved from git if not specified")
}
