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

// depsCmd represents the dependencies command
var depsCmd = &cobra.Command{
	Use:     "dependencies (REPOSITORY)",
	Aliases: []string{"deps"},
	Short:   "Retrieve the project dependencies with a given label",
	Example: `# Check if image registry/<project_name>-<image-suffix>:<tag> exists
# tag is in the format vX.Y.Z[-rcT]-N-g<short-commit-hash>
ciux get image --check <path_to_git_repository> --suffix <image_suffix>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		project, err := internal.NewProject(repositoryPath, branch, false, labelSelector)
		internal.FailOnError(err)

		for _, dep := range project.Dependencies {
			fmt.Printf("  %v\n", dep)
		}

	},
}

func init() {
	getCmd.AddCommand(depsCmd)

	util.AddLabelSelectorFlagVar(depsCmd, &labelSelector)
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
