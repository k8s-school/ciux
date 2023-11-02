/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

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
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		gitDeps, error := internal.GetDepsWorkBranch(repositoryPath)
		if error != nil {
			log.Fatal(error)
		}
		for _, gitDep := range gitDeps {
			gitDep.CloneWorkBranch()
			gitDep.Describe()
			internal.Info("Revision: %+v", gitDep.Revision)
		}
	},
}

func init() {
	getCmd.AddCommand(imageCmd)
}
