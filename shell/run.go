package shell

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kr/pty"

	"github.com/Gameye/igniter-shell-go/statemachine"
)

const outputBuffer = 200
const inputBuffer = 20
const signalBuffer = 20
const stateChangeBuffer = 20

// RunWithStateMachine runs a command
func RunWithStateMachine(
	cmd *exec.Cmd,
	config *statemachine.Config,
	withPty bool,
) (
	exit int,
	err error,
) {
	/*
		This order matters! Mostly because of the deferred closing of
		the channels. Changing this ordder might cause a panic for writing
		to a closed channel (so please, don't).
	*/
	signals := make(chan os.Signal, signalBuffer)
	defer close(signals)

	signal.Notify(signals)
	defer signal.Stop(signals)

	inputLines := make(chan string, inputBuffer)
	defer close(inputLines)

	stateChanges := make(chan statemachine.StateChange, stateChangeBuffer)
	defer close(stateChanges)

	outputLines := make(chan string, outputBuffer)
	defer close(outputLines)
	/*
	 */

	go passStateChanges(inputLines, stateChanges)
	go statemachine.Run(config, stateChanges, outputLines)

	if withPty {
		exit, err = runCommandPTY(cmd, outputLines, inputLines, signals)
		if err != nil {
			return
		}
	} else {
		exit, err = runCommand(cmd, outputLines, inputLines, signals)
		if err != nil {
			return
		}
	}

	return
}

// runCommand runs a command
func runCommand(
	cmd *exec.Cmd,
	outputLines chan<- string,
	inputLines <-chan string,
	signals <-chan os.Signal,
) (
	exit int,
	err error,
) {
	err = attachCommand(cmd, outputLines, inputLines)
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	go passSignals(cmd.Process, signals)

	exit, err = waitCommand(cmd)
	if err != nil {
		return
	}

	return
}

// runCommand runs a command as a pseudo terminal!
func runCommandPTY(
	cmd *exec.Cmd,
	outputLines chan<- string,
	inputLines <-chan string,
	signals <-chan os.Signal,
) (
	exit int,
	err error,
) {
	ptyStream, err := pty.Start(cmd)
	if err != nil {
		return
	}

	go passSignals(cmd.Process, signals)

	pipeReader, pipeWriter := io.Pipe()

	ptyTee := io.TeeReader(ptyStream, pipeWriter)

	go readLines(ptyTee, outputLines)
	go writeLines(ptyStream, inputLines)

	go io.Copy(os.Stdout, pipeReader)
	go io.Copy(ptyStream, os.Stdin)

	exit, err = waitCommand(cmd)
	if err != nil {
		return
	}

	return
}

// attachCommand attaches input and output channels to command (via pipes)
func attachCommand(
	cmd *exec.Cmd,
	outputLines chan<- string,
	inputLines <-chan string,
) (err error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	stdoutTee := io.TeeReader(stdout, stdoutPipeWriter)
	stderrTee := io.TeeReader(stderr, stderrPipeWriter)

	go readLines(stdoutTee, outputLines)
	go readLines(stderrTee, outputLines)
	go writeLines(stdin, inputLines)

	go io.Copy(os.Stdout, stdoutPipeReader)
	go io.Copy(os.Stderr, stderrPipeReader)
	go io.Copy(stdin, os.Stdin)

	return
}

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

// readLines reads lines from a reader in a channel
func readLines(
	reader io.Reader,
	lines chan<- string,
) (err error) {
	bufferedReader := bufio.NewReader(reader)

	var line string
	for {
		line, err = bufferedReader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		lines <- line
	}
}

// writeLines writes lines from a channel in a writer
func writeLines(
	writer io.Writer,
	lines <-chan string,
) (err error) {
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
) (err error) {
	var signal os.Signal
	for signal = range signals {
		err = process.Signal(signal)
		if err != nil {
			return
		}
	}
	return
}

// passStateChanges passes commands of state change structs into a
// string channel
func passStateChanges(
	inputLines chan<- string,
	stateChanges <-chan statemachine.StateChange,
) {
	for stateChange := range stateChanges {
		inputLines <- stateChange.Command
	}
}
