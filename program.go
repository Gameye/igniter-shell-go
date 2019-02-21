package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kr/pty"

	"github.com/elmerbulthuis/shell-go/statemachine"
)

const outputBuffer = 20
const inputBuffer = 20
const signalBuffer = 20
const stateChangeBuffer = 20

func main() {
	exit, err := runTf2()
	if err != nil {
		panic(err)
	}

	os.Exit(exit)
}

func runTf2() (
	exit int,
	err error,
) {
	cmd := exec.Command("docker", "run", "-ti", "docker.gameye.com/tf2")
	config := makeTf2Config()

	exit, err = runWithStateMachine(cmd, config, true)
	if err != nil {
		return
	}

	return
}

func runWithStateMachine(
	cmd *exec.Cmd,
	config *statemachine.Config,
	withPty bool,
) (
	exit int,
	err error,
) {
	signals := make(chan os.Signal, signalBuffer)
	defer close(signals)

	signal.Notify(signals)
	defer signal.Stop(signals)

	inputLines := make(chan string, inputBuffer)
	defer close(inputLines)

	stateChanges := make(chan statemachine.StateChange, stateChangeBuffer)
	defer close(stateChanges)

	go passStateChanges(inputLines, stateChanges)

	outputLines := make(chan string, outputBuffer)
	defer close(outputLines)

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

func passStateChanges(
	inputLines chan<- string,
	stateChanges <-chan statemachine.StateChange,
) {
	for stateChange := range stateChanges {
		inputLines <- stateChange.Command
	}
}

func makeTestConfig() (
	config *statemachine.Config,
) {
	config = &statemachine.Config{
		InitialState: "Off",
		States: map[string]statemachine.StateConfig{
			"On": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.TimerEventConfig{
						NextState: "Off",
						Interval:  time.Millisecond * 300,
					},
				},
			},
			"Off": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.TimerEventConfig{
						NextState: "On",
						Interval:  time.Millisecond * 300,
					},
				},
			},
		},
		Transitions: []statemachine.TransitionConfig{
			statemachine.TransitionConfig{
				To:      "On",
				Command: "echo on",
			},
			statemachine.TransitionConfig{
				To:      "Off",
				Command: "echo off",
			},
		},
	}

	return
}

func makeTf2Config() (
	config *statemachine.Config,
) {
	config = &statemachine.Config{
		InitialState: "idle",
		States: map[string]statemachine.StateConfig{
			"idle": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.LiteralEventConfig{
						NextState:  "end",
						Value:      "Unknown command \"startupmenu\"",
						IgnoreCase: true,
					},
				},
			},
			"end": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.TimerEventConfig{
						NextState: "quit",
						Interval:  time.Second * 10,
					},
				},
			},
		},
		Transitions: []statemachine.TransitionConfig{
			statemachine.TransitionConfig{
				To:      "end",
				Command: "echo Quitting",
			},
			statemachine.TransitionConfig{
				To:      "quit",
				Command: "quit",
			},
		},
	}

	return
}
