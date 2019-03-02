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
	config *statemachine.Config,
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

	// start routines

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

	signals := make(chan os.Signal, 20)
	defer close(signals)

	signal.Notify(signals, signalsToPass...)
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
	config *statemachine.Config,
) (
	exit int,
	err error,
) {
	ptyStream, ttyStream, err := pty.Open()
	if err != nil {
		return
	}
	defer ttyStream.Close()
	defer ptyStream.Close()

	cmd.Stdout = ttyStream
	cmd.Stdin = ttyStream
	cmd.Stderr = ttyStream
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	ptyTee := io.TeeReader(ptyStream, pipeWriter)

	// setup pipeline

	outputLines := readLines(ptyTee)
	stateChanges := statemachine.Run(config, outputLines)
	inputLines := selectStateCommand(stateChanges)

	// start routines

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

	signals := make(chan os.Signal, 20)
	defer close(signals)

	signal.Notify(signals, signalsToPass...)
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
