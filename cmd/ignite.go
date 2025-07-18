package cmd

import (
	"path/filepath"
	"strings"

	"github.com/k8s-school/ciux/cmd/util"
	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var branch string
var main bool

// igniteCmd represents the revision command
var igniteCmd = &cobra.Command{
	Use:     "ignite repository_path",
	Aliases: []string{"ig", "ign"},
	Short:   "Prepare integration test",
	Long: `Retrieve current revision of the repository and clone all dependencies in the correct revision.
	Check if dependencies container images are available.
	Use repository_path/.ciux configuration file to retrieve dependencies.
	Also compute the name for the container image to build:
	It uses the sourcePathes in repository_path/.ciux to retrieve the latest git commit where some code has changed,
	then it checks if and image exists in-between this commit and the current one and it returns it,
	if not it set image name for the current commit.
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		repositoryPath := internal.AbsPath(args[0])

		var project internal.Project

		project, err := internal.NewProject(repositoryPath, branch, main, labelSelector)

		internal.FailOnError(err)
		project.TemporaryRegistry = tmpRegistry
		depsBasePath := filepath.Dir(repositoryPath)

		// Retrieve dependencies sources
		err = project.RetrieveDepsSources(depsBasePath)
		internal.FailOnError(err)

		// Install dependencies Go modules
		// TODO return go modules installed and implement a print function
		goMsg, err := project.InstallGoModules()
		internal.FailOnError(err)

		// Check if dependencies container images are available
		images, err := project.CheckDepImages()
		internal.FailOnError(err)

		internal.Infof("%s", project.String())

		err = project.GetImageName(suffix, true)
		internal.FailOnError(err)
		internal.Infof("Image:\n%s", project.Image)

		goMsg = strings.TrimRight(goMsg, "\n")
		internal.Infof("Go modules installed:\n%s", goMsg)

		// Convert image to a printable string
		var imgMsg string
		for _, image := range images {
			imgMsg += "  " + image.Name() + "\n"
		}
		imgMsg = strings.TrimRight(imgMsg, "\n")
		internal.Infof("Available Images for dependencies:\n%s", imgMsg)

		// Write project configuration file
		msg, err := project.WriteOutConfig()
		internal.FailOnError(err)
		internal.Infof("%s", msg)
	},
}

func init() {
	rootCmd.AddCommand(igniteCmd)

	// Here you will define your flags and configuration settings.
	igniteCmd.Flags().BoolVarP(&main, "main", "m", false, "Only work with main project, ignore dependencies, --selector is ignored")
	igniteCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "current branch for the project, retrieved from git if not specified")
	igniteCmd.Flags().StringVarP(&suffix, "suffix", "p", "", "Suffix to add to the image name")
	igniteCmd.Flags().StringVarP(&tmpRegistry, "tmp-registry", "t", "", "Name of temporary registry used to store the image during the ci process")

	util.AddLabelSelectorFlagVar(igniteCmd, &labelSelector)
}
