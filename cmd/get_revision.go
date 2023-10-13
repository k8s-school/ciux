/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

func run(dir string) (*string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository: %v", err)
	}
	tagName, counter, headHash, dirty, err := internal.GitDescribe(*repo)
	if err != nil {
		return nil, fmt.Errorf("unable to describe commit: %v", err)
	}
	logger.Debugf("tag: %s, counter: %d, head: %s, dirty: %s", *tagName, *counter, *headHash, *dirty)
	return nil, nil
}

// revisionCmd represents the revision command
var revisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("revision called")
		dir := "/home/fjammes/src/astrolabsoftware/fink-alert-simulator"
		run(dir)
	},
}

func init() {
	getCmd.AddCommand(revisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// revisionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// revisionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
