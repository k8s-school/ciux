package internal

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warnf should be used to display a warning
func Warnf(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// LastDir returns the last element of URL path
func LastDir(permalink string) (string, error) {
	url, err := url.Parse(permalink)
	if err != nil {
		return "", err
	}
	return path.Base(url.Path), nil
}

// IsPathInSubdirectory checks if the given file path is in the given subdirectory
func IsPathInSubdirectory(filePath, subdirectory string) (bool, error) {

	// Subdirectory is root directory
	if subdirectory == "" {
		return true, nil
	}

	if filePath == "" {
		return false, fmt.Errorf("invalid arguments: filePath=%q, subdirectory=%q", filePath, subdirectory)
	}

	// Get the absolute paths
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}

	absSubdirectory, err := filepath.Abs(subdirectory)
	if err != nil {
		return false, err
	}

	// Check if the subdirectory is a prefix of the absolute file path
	return strings.HasPrefix(absFilePath, absSubdirectory+string(filepath.Separator)), nil
}

// IsPathInSubdirectories checks if the given file path is in one of the given subdirectories
// If subdirectories is empty, it returns true becuase the file path must be in the root directory
func IsPathInSubdirectories(filePath string, subdirectories []string) (bool, error) {
	if len(subdirectories) == 0 {
		return true, nil
	}
	for _, subdirectory := range subdirectories {
		isInSubdirectory, err := IsPathInSubdirectory(filePath, subdirectory)
		if err != nil {
			return false, err
		}

		if isInSubdirectory {
			return true, nil
		}
	}
	return false, nil
}
