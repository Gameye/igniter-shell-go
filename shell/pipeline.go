package shell

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/Gameye/igniter-shell-go/statemachine"
)

const outputBuffer = 200
const inputBuffer = 20
const signalBuffer = 20
const stateChangeBuffer = 20

// readLines reads lines from a reader in a channel
func readLines(
	bufferedReader *bufio.Reader,
) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)

		var err error
		var line string
		for {
			line, err = bufferedReader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			line = strings.TrimSpace(line)
			if line != "" {
				lines <- line
			}
		}
	}()

	return lines
}

// writeLines writes lines from a channel in a writer
func writeLines(
	writer io.Writer,
	lines <-chan string,
) {
	var err error
	var line string
	for line = range lines {
		_, err = io.WriteString(writer, line+"\n")
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
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
