/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/k8s-school/ciux/cmd/util"
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// configPathCmd represents the dependencies command
var configPathCmd = &cobra.Command{
	Use:     "configpath [-l <label-selector>] (REPOSITORY)",
	Aliases: []string{"cn"},
	Short:   "Get ciux configuration path",
	Example: `  ciux get configpath
  ciux get cp`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := internal.AbsPath(args[0])
		var project internal.Project

		project, err := internal.NewProject(repositoryPath, branch, main, labelSelector)

		internal.FailOnError(err)
		configPath, err := project.GetCiuxConfigFilepath()
		internal.FailOnError(err)
		// Check if the config file exists
		if !internal.FileExists(configPath) {
			internal.FailOnError(fmt.Errorf("ciux configuration file not found at %s", configPath))
		}
		fmt.Println(configPath)
	},
}

func init() {
	getCmd.AddCommand(configPathCmd)

	util.AddLabelSelectorFlagVar(configPathCmd, &labelSelector)
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
