package shell

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Gameye/igniter-shell-go/statemachine"
	"github.com/kr/pty"
)

// TODO: figure out which signals we should actually pass
var signalsToPass = []os.Signal{
	syscall.SIGABRT,
	syscall.SIGALRM,
	syscall.SIGBUS,
	syscall.SIGCHLD,
	syscall.SIGCLD,
	syscall.SIGCONT,
	syscall.SIGFPE,
	syscall.SIGHUP,
	syscall.SIGILL,
	syscall.SIGINT,
	syscall.SIGIO,
	syscall.SIGIOT,
	syscall.SIGKILL,
	syscall.SIGPIPE,
	syscall.SIGPOLL,
	syscall.SIGPROF,
	syscall.SIGPWR,
	syscall.SIGQUIT,
	syscall.SIGSEGV,
	syscall.SIGSTKFLT,
	syscall.SIGSTOP,
	syscall.SIGSYS,
	syscall.SIGTERM,
	syscall.SIGTRAP,
	syscall.SIGTSTP,
	syscall.SIGTTIN,
	syscall.SIGTTOU,
	syscall.SIGUNUSED,
	syscall.SIGURG,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
	syscall.SIGVTALRM,
	syscall.SIGWINCH,
	syscall.SIGXCPU,
	syscall.SIGXFSZ,
}

// RunWithStateMachine runs a command
func RunWithStateMachine(
	cmd *exec.Cmd,
	config *statemachine.Config,
	withPty bool,
) (
	exit int,
	err error,
) {
	signals := make(chan os.Signal, signalBuffer)
	defer close(signals)

	signal.Notify(signals, signalsToPass...)
	defer signal.Stop(signals)

	if withPty {
		exit, err = runCommandPTY(cmd, config, signals)
		if err != nil {
			return
		}
	} else {
		exit, err = runCommand(cmd, config, signals)
		if err != nil {
			return
		}
	}

	return
}

// runCommand runs a command
func runCommand(
	cmd *exec.Cmd,
	config *statemachine.Config,
	signals <-chan os.Signal,
) (
	exit int,
	err error,
) {
	// setup pipes

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

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	defer stdoutPipeReader.Close()
	defer stdoutPipeWriter.Close()
	defer stderrPipeReader.Close()
	defer stderrPipeWriter.Close()

	stdoutTee := io.TeeReader(stdout, stdoutPipeWriter)
	stderrTee := io.TeeReader(stderr, stderrPipeWriter)

	// setup pipeline

	stdoutLines := readLines(stdoutTee)
	stderrLines := readLines(stderrTee)
	outputLines := mergeLines(stdoutLines, stderrLines)

	stateChanges := statemachine.Run(config, outputLines)
	inputLines := selectStateCommand(stateChanges)

	go writeLines(stdin, inputLines)

	go io.Copy(os.Stdout, stdoutPipeReader)
	go io.Copy(os.Stderr, stderrPipeReader)
	go io.Copy(stdin, os.Stdin)

	// start the command

	err = cmd.Start()
	if err != nil {
		return
	}

	go passSignals(cmd.Process, signals)

	// wait for exit

	exit, err = waitCommand(cmd)
	if err != nil {
		return
	}

	return
}

// runCommand runs a command as a pseudo terminal!
func runCommandPTY(
	cmd *exec.Cmd,
	config *statemachine.Config,
	signals <-chan os.Signal,
) (
	exit int,
	err error,
) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	// start process and get pty

	ptyStream, err := pty.Start(cmd)
	if err != nil {
		return
	}
	defer ptyStream.Close()

	ptyTee := io.TeeReader(ptyStream, pipeWriter)

	// setup pipeline

	outputLines := readLines(ptyTee)
	stateChanges := statemachine.Run(config, outputLines)
	inputLines := selectStateCommand(stateChanges)

	go writeLines(ptyStream, inputLines)

	go io.Copy(os.Stdout, pipeReader)
	go io.Copy(ptyStream, os.Stdin)

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
