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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	igniteCmd.AddCommand(refreshCmd)
}
