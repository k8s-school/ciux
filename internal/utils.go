package internal

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

func FailOnError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Infof should be used to describe the example commands that are about to run.
func Infof(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	// Reduce string to 64 characters
}

// Warnf should be used to display a warning
func Warnf(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// LabelSelectorToFileName converts a label selector to a string which can be used in a file name
// It replaces "=" with "_" and "," with "-"
func LabelSelectorToFileName(labelSelector labels.Selector) string {
	str := labelSelector.String()
	if str == "" {
		return ""
	}
	str = strings.ReplaceAll(str, "=", "_")
	str = strings.ReplaceAll(str, ",", "-")
	str = "_" + str
	return str
}

// LastDir returns the last element of URL path
func LastDir(permalink string) (string, error) {
	url, err := url.Parse(permalink)
	if err != nil {
		return "", err
	}
	return path.Base(url.Path), nil
}

func AbsPath(path string) string {
	abspath, err := filepath.Abs(path)
	if err != nil {
		slog.Error("unable to get absolute path", "path", path, "error", err)
		os.Exit(1)
	}
	return abspath
}

// IsPathInSubdirectory checks if the given file path is in the given subdirectory
// all paths must be absolute or function will have non-deterministic behavior
func IsPathInSubdirectory(absFilePath, absSubdirectory string) (bool, error) {

	if absFilePath == "" {
		return false, fmt.Errorf("invalid argument: absFilePath=%q", absFilePath)
	}

	// Check if the subdirectory is a prefix of the absolute file path
	return strings.HasPrefix(absFilePath, absSubdirectory+string(filepath.Separator)), nil
}

// IsFileInSourcePathes checks if the given file path is in one of the given subdirectories
// If subdirectories is empty, it returns true becuase the file path must be in the root directory
func IsFileInSourcePathes(root string, filePath string, sourcePathes []string) (bool, error) {
	if len(sourcePathes) == 0 {
		return true, nil
	}

	// Get the absolute paths
	var absFilePath string
	if filepath.IsAbs(filePath) {
		return false, fmt.Errorf("invalid argument filePath must be relative: filePath=%q", filePath)
	} else {
		absFilePath = filepath.Join(root, filePath)
	}

	for _, sourcePath := range sourcePathes {

		var absSourcePath string
		if filepath.IsAbs(sourcePath) {
			return false, fmt.Errorf("invalid argument sourcePath must be relative: sourcePath=%q", sourcePath)
		} else {
			absSourcePath = filepath.Join(root, sourcePath)
		}

		isDir, err := isDirectory(absSourcePath)
		if err != nil {
			return false, err
		}

		if isDir {
			isInSubdirectory, err := IsPathInSubdirectory(absFilePath, absSourcePath)
			if err != nil {
				return false, err
			}

			if isInSubdirectory {
				return true, nil
			}
		} else {
			slog.Debug("Check if file path is equal to source path", "filePath", filePath, "sourcePath", sourcePath)
			if sourcePath == filePath {
				return true, nil
			}
		}
	}
	return false, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
