/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/k8s-school/ciux/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func run(dir string, deps []string) (*string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository: %v", err)
	}
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXX" + dir)
	tagName, counter, headHash, dirty, err := internal.GitDescribe(*repo)
	if err != nil {
		return nil, fmt.Errorf("unable to describe commit: %v", err)
	}
	branchName, err := internal.GitBranchName(*repo)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve branch name: %v", err)
	}
	logger.Debugf("tag: %s, counter: %d, head: %s, dirty: %t", *tagName, *counter, *headHash, *dirty)
	logger.Debugf("branch: %s", *branchName)

	// Test if branchName equals master or main
	if *branchName == "master" || *branchName == "main" {
		log.Debug().Msg("Branch is master or main")
	}

	for _, dep := range deps {

		_, err := git.PlainClone("/tmp/foo", false, &git.CloneOptions{
			URL:          dep,
			Progress:     os.Stdout,
			SingleBranch: true,
			//ReferenceName: *branchName,
		})

		if err != nil {
			return nil, fmt.Errorf("unable to open git repository: %v", err)
		}
	}
	return nil, nil
}

// revisionCmd represents the revision command
var revisionCmd = &cobra.Command{
	Use:   "revision (REPOSITORY) (DEPENDENCY_REPOSITORIES...)",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("revision called")
		repository := args[0]
		_, err := run(repository, args[1:])
		internal.CheckIfError(err)
	},
}

func init() {
	getCmd.AddCommand(revisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// revisionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
