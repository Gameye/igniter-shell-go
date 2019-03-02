package shell

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/Gameye/igniter-shell-go/statemachine"
)

// readLines reads lines from a reader in a channel
func readLines(
	reader io.Reader,
) <-chan string {
	lines := make(chan string)
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
		// FIX: read /dev/ptmx: input/output error
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