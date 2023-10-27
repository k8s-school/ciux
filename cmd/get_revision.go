/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var dependency string

// revisionCmd represents the revision command
var revisionCmd = &cobra.Command{
	Use:   "revision (REPOSITORY) (DEPENDENCY_REPOSITORIES...)",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repository := args[0]
		dependencies := args[1:]
		gitMeta := internal.GitMeta{}
		err := gitMeta.Analyze(repository, dependencies)
		internal.CheckIfError(err)
		if len(dependencies) == 0 {
			internal.Info("Version: %+v", gitMeta.Revision)
			return
		}

		for _, dep := range dependencies {
			depGitMeta, e := internal.GitLsRemote(dep)
			internal.CheckIfError(e)
			internal.Info("Branches: %+v", depGitMeta.Branches)
			internal.Info("Depedenncy version: %+v", depGitMeta.Revision)
		}
	},
}

func init() {
	getCmd.AddCommand(revisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	revisionCmd.Flags().StringVarP(&dependency, "dependency", "d", "", "Dependency repository")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
