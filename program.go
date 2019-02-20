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

	"github.com/elmerbulthuis/shell-go/statemachine"
)

const outputBuffer = 20
const inputBuffer = 20
const signalBuffer = 20
const stateChangeBuffer = 20

func main() {
	exit, err := run()
	if err != nil {
		panic(err)
	}

	os.Exit(exit)
}

func run() (
	exit int,
	err error,
) {
	cmd := exec.Command("bash")
	config := makeTestConfig()

	exit, err = runWithStateMachine(cmd, config)
	if err != nil {
		return
	}

	return
}

func runWithStateMachine(
	cmd *exec.Cmd,
	config *statemachine.Config,
) (
	exit int,
	err error,
) {
	outputLines := make(chan string, outputBuffer)
	defer close(outputLines)

	inputLines := make(chan string, inputBuffer)
	defer close(inputLines)

	stateChanges := make(chan statemachine.StateChange, stateChangeBuffer)
	defer close(stateChanges)

	err = attachCommand(cmd, outputLines, inputLines)
	if err != nil {
		return
	}

	go passStateChange(inputLines, stateChanges)
	go statemachine.Run(config, stateChanges, outputLines)

	exit, err = runCommand(cmd)
	if err != nil {
		return
	}

	return
}

func runCommand(
	cmd *exec.Cmd,
) (
	exit int,
	err error,
) {
	signals := make(chan os.Signal, signalBuffer)
	defer close(signals)

	signal.Notify(signals)
	defer signal.Stop(signals)

	err = cmd.Start()
	if err != nil {
		return
	}

	go passSignals(cmd.Process, signals)

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

	stdoutBufferedReader := bufio.NewReader(io.TeeReader(stdout, stdoutPipeWriter))
	stderrBufferedReader := bufio.NewReader(io.TeeReader(stderr, stderrPipeWriter))

	go readLines(stdoutBufferedReader, outputLines)
	go readLines(stderrBufferedReader, outputLines)
	go writeLines(stdin, inputLines)

	go io.Copy(os.Stdout, stdoutPipeReader)
	go io.Copy(os.Stderr, stderrPipeReader)
	go io.Copy(stdin, os.Stdin)

	return
}

func readLines(
	reader *bufio.Reader,
	lines chan<- string,
) (err error) {
	var line string
	for {
		line, err = reader.ReadString('\n')
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

func passStateChange(
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
						Interval:  time.Second * 1,
					},
				},
			},
			"Off": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.TimerEventConfig{
						NextState: "On",
						Interval:  time.Second * 1,
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
