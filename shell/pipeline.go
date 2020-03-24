package shell

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/Gameye/igniter-shell-go/runner"
	"github.com/Gameye/igniter-shell-go/utils"
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

			line = utils.StripSpecial(line)

			line = strings.TrimSpace(line)
			if line != "" {
				lines <- line
			}
		}
		/*
			ignore possible read /dev/ptmx: input/output error here
			also ignore possible bufio.Scanner: token too long (this
			is a problem)

			err := scanner.Err()
			if err != nil {
				panic(err)
			}
		*/
	}()

	return lines
}

func selectStateCommand(
	stateChanges <-chan runner.StateChange,
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
