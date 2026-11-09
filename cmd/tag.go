/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"strings"

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

		tagArgs := []string{"tag", "-a", "-m", "Release " + newTag, newTag}
		// Fully qualified refspec: a branch with the same name is not pushed,
		// and --no-follow-tags pushes no other tag, whatever the git config
		pushArgs := []string{"push", "--no-follow-tags", "origin", "refs/tags/" + newTag}
		if dryRun {
			fmt.Println(shellJoin(append([]string{"git", "-C", repositoryPath}, tagArgs...)))
			fmt.Println(shellJoin(append([]string{"git", "-C", repositoryPath}, pushArgs...)))
			return
		}
		internal.FailOnError(runGit(repositoryPath, tagArgs...))
		if err := runGit(repositoryPath, pushArgs...); err != nil {
			// Remove the local tag, so that the command can be run again
			if delErr := runGit(repositoryPath, "tag", "-d", newTag); delErr != nil {
				err = fmt.Errorf("%v\nunable to delete local tag %s: %v", err, newTag, delErr)
			}
			internal.FailOnError(err)
		}
		internal.Infof("Tag %s created and pushed", newTag)
	},
}

// runGit runs git in repositoryPath, without a shell, its error includes
// the git output
func runGit(repositoryPath string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", repositoryPath}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

// shellJoin returns args as a command line which can be pasted in a shell
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
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
