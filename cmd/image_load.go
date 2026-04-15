/*
Copyright © 2025 Fabrice Jammes fabrice.jammes@in2p3.fr

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/k8s-school/ciux/internal"
	"github.com/spf13/cobra"
)

var (
	runnerType   string
	imageUrl     string
	buildStatus  string
	artifactPath string
	maxAge       string
	imagePattern string
)

// imageLoadCmd represents the image load command
var imageLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load container image based on runner type and build status",
	Long: `Load a container image in the appropriate way based on the runner type (self-hosted vs GitHub Actions)
and the build status (locally built vs existing remote image).

For self-hosted runners with locally built images, it verifies the image exists in the local Docker registry.
For GitHub Actions with locally built images, it loads the image from artifacts.
For existing remote images, it assumes the image is already available.`,
	Example: `# Load image in GitHub Actions workflow
ciux image load --runner "ubuntu-latest" --image-url "registry.io/project:v1.0.0" --build-status "true"

# Load image on self-hosted runner
ciux image load --runner "['self-hosted']" --image-url "registry.io/project:v1.0.0" --build-status "false"`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runImageLoad()
		internal.FailOnError(err)
	},
}

// imageManagementCmd represents the image parent command
var imageManagementCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage container images for CI/CD workflows",
	Long: `A collection of commands to manage container images in CI/CD workflows,
particularly for handling differences between self-hosted and GitHub Actions runners.`,
}

// imageCleanupCmd represents the image cleanup command
var imageCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old container images",
	Long: `Remove container images older than the specified age.
This is particularly useful for self-hosted runners to prevent disk space accumulation.

The command removes:
1. Images matching the specified pattern that are older than max-age
2. Dangling images (images not tagged or referenced)

For safety, only images matching the pattern are removed, not all images.`,
	Example: `# Clean up fink-broker images older than 5 days
ciux image cleanup --max-age 5d --pattern fink-broker

# Clean up all images matching pattern older than 1 week
ciux image cleanup --max-age 1w --pattern "gitlab-registry.in2p3.fr/astrolabsoftware"`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runImageCleanup()
		internal.FailOnError(err)
	},
}

func init() {
	rootCmd.AddCommand(imageManagementCmd)
	imageManagementCmd.AddCommand(imageLoadCmd)
	imageManagementCmd.AddCommand(imageCleanupCmd)

	// Flags for the load subcommand
	imageLoadCmd.Flags().StringVarP(&runnerType, "runner", "r", "", "Runner type (e.g., 'ubuntu-latest' or \"['self-hosted']\")")
	imageLoadCmd.Flags().StringVarP(&imageUrl, "image-url", "i", "", "Container image URL")
	imageLoadCmd.Flags().StringVarP(&buildStatus, "build-status", "b", "", "Build status ('true' if locally built, 'false' if existing)")
	imageLoadCmd.Flags().StringVarP(&artifactPath, "artifact-path", "a", "artifacts/image.tar", "Path to artifact file for GitHub Actions")

	imageLoadCmd.MarkFlagRequired("runner")
	imageLoadCmd.MarkFlagRequired("image-url")
	imageLoadCmd.MarkFlagRequired("build-status")

	// Flags for the cleanup subcommand
	imageCleanupCmd.Flags().StringVarP(&maxAge, "max-age", "m", "5d", "Maximum age for images (e.g., '5d', '1w', '2h')")
	imageCleanupCmd.Flags().StringVarP(&imagePattern, "pattern", "p", "", "Pattern to match image names (required)")

	imageCleanupCmd.MarkFlagRequired("pattern")
}

func runImageLoad() error {
	internal.Infof("Loading image: %s (build-status: %s, runner: %s)", imageUrl, buildStatus, runnerType)

	// Parse build status
	isLocallyBuilt := strings.ToLower(buildStatus) == "true"

	// Determine if this is a self-hosted runner
	isSelfHosted := strings.Contains(runnerType, "self-hosted")

	if !isLocallyBuilt {
		// Image already exists (remote), nothing to load
		internal.Infof("Using existing remote image: %s", imageUrl)
		return nil
	}

	// Image was built locally, handle based on runner type
	if isSelfHosted {
		return loadImageSelfHosted()
	} else {
		return loadImageGitHubActions()
	}
}

func loadImageSelfHosted() error {
	internal.Infof("Self-hosted runner: verifying image %s in local Docker registry", imageUrl)

	// Check if image exists in local Docker registry
	cmd := exec.Command("docker", "images", "--filter", fmt.Sprintf("reference=%s", imageUrl), "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Docker images: %v", err)
	}

	// Check if our image is in the output
	if strings.Contains(string(output), imageUrl) {
		internal.Infof("✓ Image %s found in local Docker registry", imageUrl)

		// Show image details for verification
		showCmd := exec.Command("docker", "images", "--filter", fmt.Sprintf("reference=%s", imageUrl))
		showOutput, err := showCmd.Output()
		if err != nil {
			internal.Warnf("Could not show image details: %v", err)
		} else {
			fmt.Print(string(showOutput))
		}

		return nil
	}

	// Image not found - this is an error
	internal.Infof("✗ Image %s not found in local Docker registry", imageUrl)
	internal.Infof("Available images:")

	listCmd := exec.Command("docker", "images")
	listOutput, err := listCmd.Output()
	if err != nil {
		internal.Warnf("Could not list Docker images: %v", err)
	} else {
		fmt.Print(string(listOutput))
	}

	return fmt.Errorf("image %s not found in local Docker registry", imageUrl)
}

func loadImageGitHubActions() error {
	internal.Infof("GitHub Actions runner: loading image from artifact %s", artifactPath)

	// Check if artifact file exists
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact file %s not found", artifactPath)
	}

	// Load image from tar archive
	cmd := exec.Command("docker", "load", "--input", artifactPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load image from %s: %v\nOutput: %s", artifactPath, err, string(output))
	}

	internal.Infof("✓ Successfully loaded image from %s", artifactPath)
	fmt.Print(string(output))

	return nil
}

func runImageCleanup() error {
	internal.Infof("Cleaning up images matching pattern '%s' older than %s", imagePattern, maxAge)

	// Parse max age duration
	duration, err := parseMaxAge(maxAge)
	if err != nil {
		return fmt.Errorf("invalid max-age format: %v", err)
	}

	cutoffTime := time.Now().Add(-duration)
	internal.Infof("Removing images created before: %s", cutoffTime.Format("2006-01-02 15:04:05"))

	// Get list of images with their creation time
	cmd := exec.Command("docker", "images", "--format", "table {{.Repository}}\t{{.Tag}}\t{{.CreatedAt}}\t{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list Docker images: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		internal.Infof("No images found")
		return nil
	}

	// Skip header line
	var removedCount int
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		repo := parts[0]
		tag := parts[1]
		createdStr := strings.Join(parts[2:len(parts)-1], " ") // Created time might have spaces
		imageId := parts[len(parts)-1]

		// Check if image matches pattern
		imageName := repo + ":" + tag
		if !strings.Contains(imageName, imagePattern) {
			continue
		}

		// Parse creation time
		createdTime, err := parseDockerTime(createdStr)
		if err != nil {
			internal.Warnf("Could not parse creation time '%s' for image %s: %v", createdStr, imageName, err)
			continue
		}

		if createdTime.Before(cutoffTime) {
			internal.Infof("Removing old image: %s (created: %s)", imageName, createdStr)
			removeCmd := exec.Command("docker", "rmi", imageId)
			if removeOutput, err := removeCmd.CombinedOutput(); err != nil {
				internal.Warnf("Failed to remove image %s: %v\nOutput: %s", imageId, err, string(removeOutput))
			} else {
				internal.Infof("✓ Removed image %s", imageName)
				removedCount++
			}
		} else {
			internal.Infof("Keeping recent image: %s (created: %s)", imageName, createdStr)
		}
	}

	// Clean up dangling images
	internal.Infof("Cleaning up dangling images")
	pruneCmd := exec.Command("docker", "image", "prune", "-f")
	pruneOutput, err := pruneCmd.CombinedOutput()
	if err != nil {
		internal.Warnf("Failed to prune dangling images: %v\nOutput: %s", err, string(pruneOutput))
	} else {
		internal.Infof("✓ Cleaned up dangling images")
		fmt.Print(string(pruneOutput))
	}

	// Show remaining images matching pattern
	internal.Infof("Remaining images matching pattern '%s':", imagePattern)
	listCmd := exec.Command("docker", "images")
	listOutput, err := listCmd.Output()
	if err != nil {
		internal.Warnf("Could not list remaining images: %v", err)
	} else {
		found := false
		for _, line := range strings.Split(string(listOutput), "\n") {
			if strings.Contains(line, imagePattern) {
				if !found {
					fmt.Println("IMAGE\t\t\t\tTAG\t\tIMAGE ID\t\tCREATED\t\tSIZE")
					found = true
				}
				fmt.Println(line)
			}
		}
		if !found {
			internal.Infof("No images matching pattern '%s' found", imagePattern)
		}
	}

	internal.Infof("Cleanup completed: removed %d images", removedCount)
	return nil
}

func parseMaxAge(maxAge string) (time.Duration, error) {
	if len(maxAge) < 2 {
		return 0, fmt.Errorf("invalid format")
	}

	valueStr := maxAge[:len(maxAge)-1]
	unit := maxAge[len(maxAge)-1:]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value")
	}

	switch unit {
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	default:
		return 0, fmt.Errorf("unsupported unit '%s' (use d, h, w, m)", unit)
	}
}

func parseDockerTime(timeStr string) (time.Time, error) {
	// Docker time formats can vary, try common formats
	formats := []string{
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2 days ago",
		"3 weeks ago",
		"1 month ago",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// Handle relative time strings
	if strings.Contains(timeStr, "ago") {
		return parseDockerRelativeTime(timeStr)
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", timeStr)
}

func parseDockerRelativeTime(timeStr string) (time.Time, error) {
	now := time.Now()

	if strings.Contains(timeStr, "second") {
		// Extract number of seconds
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * time.Second), nil
			}
		}
	}

	if strings.Contains(timeStr, "minute") {
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * time.Minute), nil
			}
		}
	}

	if strings.Contains(timeStr, "hour") {
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * time.Hour), nil
			}
		}
	}

	if strings.Contains(timeStr, "day") {
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * 24 * time.Hour), nil
			}
		}
	}

	if strings.Contains(timeStr, "week") {
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * 7 * 24 * time.Hour), nil
			}
		}
	}

	if strings.Contains(timeStr, "month") {
		fields := strings.Fields(timeStr)
		if len(fields) >= 1 {
			if value, err := strconv.Atoi(fields[0]); err == nil {
				return now.Add(-time.Duration(value) * 30 * 24 * time.Hour), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("unsupported relative time: %s", timeStr)
}