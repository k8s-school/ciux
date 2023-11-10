/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		msg, err := internal.String(repositoryPath)
		internal.FailOnError(err)
		internal.Info(msg)

		depsDir := filepath.Dir(repositoryPath)

		// Clone dependencies directories and checkout the correct revision
		// Check container images exist
		err = project.ScanRemoteDeps()
		internal.FailOnError(err)
		for _, gitDep := range project.GitDeps {
			singleBranch := true
			gitDep.Clone(depsDir, singleBranch)
			rev, err := gitDep.GetRevision()
			internal.FailOnError(err)
			internal.Info("Dep repo: %s, version: %+v", gitDep.Url, rev.GetVersion())
			// TODO: Set image path at configuration time
			depName, err := internal.LastDir(gitDep.Url)
			internal.FailOnError(err)
			imageUrl := fmt.Sprintf("%s/%s:%s", project.Config.Registry, depName, rev.GetVersion())
			_, ref, err := internal.DescImage(imageUrl)
			internal.FailOnError(err)
			internal.Info("Image ref: %s", ref)
		}
		err = project.WriteOutConfig()
		internal.FailOnError(err)
	},
}

func init() {
	rootCmd.AddCommand(igniteCmd)
}
