package shell

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

// waitCommand waits for a command to exit and returns the exit code
func waitCommand(
	cmd *exec.Cmd,
) (
	exit int,
	err error,
) {
	err = cmd.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			err = nil
			exit = status.ExitStatus()
			return
		}
	}

	if err != nil {
		return
	}

	return
}

// passLines writes lines from a channel in a writer
func passLines(
	writer io.Writer,
	lines <-chan string,
) (
	err error,
) {
	var line string
	for line = range lines {
		_, err = io.WriteString(writer, line+"\n")
		if err != nil {
			return
		}
	}
	return
}

// passSignals passes singnals from a channel to a process
func passSignals(
	process *os.Process,
	signals <-chan os.Signal,
) {
	var err error
	var signal os.Signal
	for signal = range signals {
		err = process.Signal(signal)
		if err != nil {
			return
		}
	}
	return
}
