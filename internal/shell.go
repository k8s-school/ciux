package internal

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

const shell = "bash"

func ExecCmd(command string, interactive bool, dryRun bool) (string, string, error) {

	var stdoutBuf, stderrBuf bytes.Buffer
	if !dryRun {
		slog.Info("Run in shell:", "command", command)
		cmd := exec.Command(shell, "-c", command)
		if interactive {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
			cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
		}
		err := cmd.Run()
		if err != nil {
			return "", "", fmt.Errorf("failed to run command %s in shell: %s", command, err)
		}
		// logger.Infof("stdout %v", stdoutBuf)
		// logger.Infof("stderr %v", stderrBuf)

	} else {
		slog.Info("Dry run:", "command", command)
	}
	return stdoutBuf.String(), stderrBuf.String(), nil
}
