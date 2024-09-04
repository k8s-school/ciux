/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/k8s-school/ciux/cmd/util"
	"github.com/k8s-school/ciux/internal" // Add this line to import the internal package
	"github.com/spf13/cobra"
)

// clusterNameCmd represents the dependencies command
var clusterNameCmd = &cobra.Command{
	Use:     "clustername (REPOSITORY)",
	Aliases: []string{"cn"},
	Short:   "Get a kind cluster name based on username and branch",
	Example: `  ciux get clustername
  ciux get cn`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := args[0]
		gitMeta, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)
		rev, err := gitMeta.GetHeadRevision()
		internal.FailOnError(err)

		// Get USER env variable
		user := os.Getenv("USER")
		if user == "" {
			user = "unknown"
		}

		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			log.Fatal(err)
		}

		dirty := ""
		if rev.Dirty {
			dirty = "-dirty"
		}
		clusterName := fmt.Sprintf("%s-%s-%.6s%s", user, rev.Branch, rev.Hash, dirty)
		clusterName = reg.ReplaceAllString(clusterName, "-")
		clusterName = strings.ToLower(clusterName)

		// "getconf HOST_NAME_MAX" returns usually 64
		internal.Infof("%.50s", clusterName)

	},
}

func init() {
	getCmd.AddCommand(clusterNameCmd)

	util.AddLabelSelectorFlagVar(clusterNameCmd, &labelSelector)
}

// Create a golang function which returns the revision of a git repository
// Path: cmd/get_revision.go

// Create a golang function which returns the tag of a git repository
// Path: cmd/get_tag.go
