/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// releaseCmd represents the revision command
var releaseCmd = &cobra.Command{
	Use:     "release repository_path",
	Aliases: []string{"rel"},
	Short:   "Prepare a release",
	Long:    `Prepare a release`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		project := internal.NewProject(repositoryPath)
		msg, err := internal.String(repositoryPath)
		internal.FailOnError(err)
		internal.Info(msg)

		err = project.GetDepsWorkBranch()
		internal.FailOnError(err)
		for _, gitDep := range project.GitDeps {
			singleBranch := true
			gitDep.Clone("", singleBranch)
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
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
