package shell

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/Gameye/igniter-shell-go/statemachine"
)

// readLines reads lines from a reader in a channel
func readLines(
	reader io.Reader,
) <-chan string {
	lines := make(chan string, 10000) // yes we need a big buffer
	scanner := bufio.NewScanner(reader)

	go func() {
		defer close(lines)

		var line string
		for scanner.Scan() {
			line = scanner.Text()

			line = strings.TrimSpace(line)
			if line != "" {
				lines <- line
			}
		}
		err := scanner.Err()
		if err != nil {
			panic(err)
		}
	}()

	return lines
}

func selectStateCommand(
	stateChanges <-chan statemachine.StateChange,
) <-chan string {
	stateCommands := make(chan string)

	go func() {
		defer close(stateCommands)

		for stateChange := range stateChanges {
			stateCommands <- stateChange.Command
		}
	}()

	return stateCommands
}

func mergeLines(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	wg.Add(len(cs))
	output := func(c <-chan string) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
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
