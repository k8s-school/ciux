/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// envCmd represents the refresh command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print ciux environment variables",
	Long:  `Print ciux environment variables for current local dependencies.`,
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
	igniteCmd.AddCommand(envCmd)
}
