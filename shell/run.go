package shell

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Gameye/igniter-shell-go/runner"
	"github.com/kr/pty"
)

// RunWithRunner runs a command
func RunWithRunner(
	cmd *exec.Cmd,
	config *runner.Config,
	withPty bool,
) (
	exit int,
	err error,
) {

	if withPty {
		exit, err = runCommandPTY(cmd, config)
		if err != nil {
			return
		}
	} else {
		exit, err = runCommand(cmd, config)
		if err != nil {
			return
		}
	}

	return
}

// runCommand runs a command
func runCommand(
	cmd *exec.Cmd,
	config *runner.Config,
) (
	exit int,
	err error,
) {
	// setup pipes

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	defer stdoutPipeReader.Close()
	defer stdoutPipeWriter.Close()
	defer stderrPipeReader.Close()
	defer stderrPipeWriter.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	defer stderr.Close()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	defer stdin.Close()

	stdoutTee := io.TeeReader(stdout, stdoutPipeWriter)
	stderrTee := io.TeeReader(stderr, stderrPipeWriter)

	// setup pipeline

	stdoutLines := readLines(stdoutTee)
	stderrLines := readLines(stderrTee)
	outputLines := mergeLines(stdoutLines, stderrLines)

	stateChanges := runner.Run(config, outputLines)

	inputLines := make(chan string)
	defer close(inputLines)
	signals := make(chan os.Signal, 20)
	defer close(signals)

	// start routines

	go func() {
		for stateChangeUnknown := range stateChanges {
			switch stateChange := stateChangeUnknown.(type) {
			case runner.CommandStateChange:
				inputLines <- stateChange.Command

			case runner.ExitStateChange:
				signals <- os.Interrupt
			}
		}
	}()

	go func() {
		var err error
		_, err = io.Copy(os.Stdout, stdoutPipeReader)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		var err error
		_, err = io.Copy(os.Stderr, stderrPipeReader)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		var err error
		_, err = io.Copy(stdin, os.Stdin)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		var err error
		err = passLines(stdin, inputLines)
		if err != nil {
			panic(err)
		}
	}()

	// start the command

	err = cmd.Start()
	if err != nil {
		return
	}

	signal.Notify(signals)
	defer signal.Stop(signals)

	go passSignals(cmd.Process, signals)

	// wait for exit

	exit, err = waitCommand(cmd)
	if err != nil {
		return
	}

	return
}

// runCommand runs a command in a pseudo terminal!
func runCommandPTY(
	cmd *exec.Cmd,
	config *runner.Config,
) (
	exit int,
	err error,
) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	ptyStream, ttyStream, err := pty.Open()
	if err != nil {
		return
	}
	defer ptyStream.Close()
	defer ttyStream.Close()

	cmd.Stdout = ttyStream
	cmd.Stdin = ttyStream
	cmd.Stderr = ttyStream
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	ptyTee := io.TeeReader(ptyStream, pipeWriter)

	// setup pipeline

	outputLines := readLines(ptyTee)
	stateChanges := runner.Run(config, outputLines)

	inputLines := make(chan string)
	defer close(inputLines)
	signals := make(chan os.Signal, 20)
	defer close(signals)

	// start routines

	go func() {
		for stateChangeUnknown := range stateChanges {
			switch stateChange := stateChangeUnknown.(type) {
			case runner.CommandStateChange:
				inputLines <- stateChange.Command

			case runner.ExitStateChange:
				signals <- os.Interrupt
			}
		}
	}()

	go func() {
		var err error
		_, err = io.Copy(os.Stdout, pipeReader)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		var err error
		_, err = io.Copy(ptyStream, os.Stdin)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		var err error
		err = passLines(ptyStream, inputLines)
		if err != nil {
			panic(err)
		}
	}()

	// start the command

	err = cmd.Start()
	if err != nil {
		return
	}

	signal.Notify(signals)
	defer signal.Stop(signals)

	go passSignals(cmd.Process, signals)

	// wait for exit

	exit, err = waitCommand(cmd)
	if err != nil {
		return
	}

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
	var signal os.Signal
	for signal = range signals {
		/*
			ignore error, possible errors include the process to be stopped
			or not started yet.
		*/
		_ = process.Signal(signal)
	}
	return
}
