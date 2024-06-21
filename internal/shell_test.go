package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecCmd(t *testing.T) {
	require := require.New(t)

	// Test case 1: Command execution succeeds
	command := "echo 'Hello, World!'"
	stdout, stderr, err := ExecCmd(command, false)
	require.NoError(err)
	require.Equal("Hello, World!\n", stdout)
	require.Equal("", stderr)

	// Test case 2: Command execution fails
	command = "nonexistent-command"
	stdout, stderr, err = ExecCmd(command, false)
	require.Error(err)
	require.Equal("", stdout)
	require.Equal("bash: line 1: nonexistent-command: command not found\n", stderr)

	// Test case 3: Dry run mode
	command = "echo 'Hello, World!'"
	stdout, stderr, err = ExecCmd(command, true)
	require.NoError(err)
	require.Equal("", stdout)
	require.Equal("", stderr)
}
