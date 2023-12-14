/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh ciux configuration file",
	Long: `Refresh ciux configuration file using only current local dependencies.

No remote dependencies are used.`,
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		project, err := internal.NewProject(repositoryPath, branch, "")
		internal.FailOnError(err)
		depsBasePath := filepath.Dir(repositoryPath)
		err = project.AddInPlaceDepsSources(depsBasePath)
		internal.FailOnError(err)

		msg, err := project.WriteOutConfig()
		internal.FailOnError(err)
		// Use 'refresh' in output message
		internal.Infof("%s", msg)
	},
}

func init() {
	igniteCmd.AddCommand(refreshCmd)
}
