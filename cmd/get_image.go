/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// imageCmd represents the revision command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		internal.ReadConfig(repositoryPath)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		gitMeta, err := internal.GitOpen(repositoryPath)
		cobra.CheckErr(err)

		err = gitMeta.Describe()
		cobra.CheckErr(err)
		internal.Info("Version: %+v", gitMeta.Revision)

		c := internal.GetConfig()
		for _, dep := range c.Dependencies {
			depGit := &internal.Git{IsRemote: true, Url: dep.Url}
			hasBranch, err := depGit.HasBranch(gitMeta.Revision.Branch)
			cobra.CheckErr(err)
			internal.Info("Dependency -> hasBranch: %t", hasBranch)
		}
	},
}

func init() {
	getCmd.AddCommand(imageCmd)
}
