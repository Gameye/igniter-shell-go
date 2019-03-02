package shell

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/Gameye/igniter-shell-go/statemachine"
)

const outputBuffer = 200
const inputBuffer = 20
const signalBuffer = 20
const stateChangeBuffer = 20

// readLines reads lines from a reader in a channel
func readLines(
	bufferedReader *bufio.Reader,
	lines chan<- string,
) {
	var err error
	var line string
	for {
		line, err = bufferedReader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}

		line = strings.TrimSpace(line)
		if line != "" {
			lines <- line
		}
	}
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
