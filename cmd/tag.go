/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:     "tag repository_path",
	Aliases: []string{"rel"},
	Short:   "Create a versioned tag for a git repository and push it",
	Long: `Create the next release candidate tag on the HEAD of a git repository,
and push it to the origin remote. With --dry-run, only print the commands.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repositoryPath := internal.AbsPath(args[0])
		git, err := internal.NewGit(repositoryPath)
		internal.FailOnError(err)

		rev, err := git.GetHeadRevision()
		internal.FailOnError(err)
		internal.FailOnError(checkTaggable(rev))
		newTag, err := rev.UpgradeTag()
		internal.FailOnError(err)

		gitCmd := fmt.Sprintf("git -C '%s'", repositoryPath)
		tagCmd := fmt.Sprintf("%s tag -a -m 'Release %s' '%s'", gitCmd, newTag, newTag)
		pushCmd := fmt.Sprintf("%s push origin '%s'", gitCmd, newTag)
		if dryRun {
			fmt.Println(tagCmd)
			fmt.Println(pushCmd)
			return
		}
		internal.FailOnError(run(tagCmd))
		if err := run(pushCmd); err != nil {
			// Remove the local tag, so that the command can be run again
			if delErr := run(fmt.Sprintf("%s tag -d '%s'", gitCmd, newTag)); delErr != nil {
				err = fmt.Errorf("%v\nunable to delete local tag %s: %v", err, newTag, delErr)
			}
			internal.FailOnError(err)
		}
		internal.Infof("Tag %s created and pushed", newTag)
	},
}

// run runs a shell command, its error includes the command stderr
func run(command string) error {
	_, stderr, err := internal.ExecCmd(command, false)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}
	return nil
}

// checkTaggable returns an error if HEAD must not get a new release tag:
// it has uncommitted changes, or it is already tagged
func checkTaggable(rev *internal.GitRevision) error {
	if rev.Dirty {
		return fmt.Errorf("repository has uncommitted changes, commit them before tagging")
	}
	if rev.Tag != "" && rev.Counter == 0 {
		return fmt.Errorf("HEAD is already tagged %s", rev.Tag)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
