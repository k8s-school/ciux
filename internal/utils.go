package internal

import (
	"fmt"
	"net/url"
	"os"
	"path"
)

func FailOnError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// lastDir returns the last element of URL path
func lastDir(permalink string) (string, error) {
	url, err := url.Parse(permalink)
	if err != nil {
		return "", err
	}
	return path.Base(url.Path), nil
}
