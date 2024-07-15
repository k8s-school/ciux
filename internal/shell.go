package internal

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
)

const shell = "bash"

func ExecCmd(command string, dryRun bool) (string, string, error) {

	var outErr error
	stderrBuf := new(bytes.Buffer)
	stdoutBuf := new(bytes.Buffer)
	if !dryRun {
		cmd := exec.Command(shell, "-c", command)
		cmd.Stdout = stdoutBuf
		cmd.Stderr = stderrBuf
		err := cmd.Run()
		if err != nil {
			outErr = fmt.Errorf("failed to run command %s in shell: %s", command, err)
		}
		// logger.Infof("stdout %v", stdoutBuf)
		// logger.Infof("stderr %v", stderrBuf)

	} else {
		slog.Info("Dry run:", "command", command)
	}
	return stdoutBuf.String(), stderrBuf.String(), outErr
}
