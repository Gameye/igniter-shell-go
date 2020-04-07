package runner

import (
	"os"
	"strings"
	"time"
)

/*
StateChange contains information on a changes state
*/
type StateChange interface{}

/*
CommandStateChange sends commands to the process
*/
type CommandStateChange struct {
	NextState string
	Command   string
}

/*
SignalStateChange sends a signal to the process
*/
type SignalStateChange struct {
	NextState string
	Signal    os.Signal
}

/*
KillStateChange kills the process
*/
type KillStateChange struct {
	NextState string
}

/*
Run runs a new Runner
*/
func Run(
	config *Config,
	actionChannel <-chan string,
) <-chan StateChange {
	changeChannel := make(chan StateChange)

	go func() {
		defer close(changeChannel)

		state := config.InitialState
		for {
			start := time.Now()

			// find state config
			stateConfig, hasStateConfig := config.States[state]
			if !hasStateConfig {
				return
			}

			// setup timer event
			interval := time.Duration(1<<63 - 1) // maxDuration from time.go:624
			for _, eventConfigObject := range stateConfig.Events {
				switch eventConfig := eventConfigObject.(type) {

				case TimerEventConfig:
					// we only use the first timer
					if time.Duration(eventConfig.Interval) < interval {
						interval = time.Duration(eventConfig.Interval)
					}
				}
			}
			timer := time.NewTimer(interval)

			nextState := ""
			for nextState == "" {
				select {
				case now := <-timer.C:
					for _, eventConfigObject := range stateConfig.Events {
						switch eventConfig := eventConfigObject.(type) {

						case TimerEventConfig:
							nextState = handleTimerEvent(&eventConfig, now.Sub(start))
							if nextState != "" {
								break
							}
						}
					}

				case action, more := <-actionChannel:
					if !more {
						return
					}

					action = strings.TrimSpace(action)

				loop:
					for _, eventConfigObject := range stateConfig.Events {
						switch eventConfig := eventConfigObject.(type) {

						case LiteralEventConfig:
							nextState = handleLiteralEvent(&eventConfig, action)
							if nextState != "" {
								break loop
							}

						case RegexEventConfig:
							nextState = handleRegexEvent(&eventConfig, action)
							if nextState != "" {
								break loop
							}

						}
					}

				}

			}

			timer.Stop()

			if nextState != state {
				pushState(
					nextState,
					state,
					config,
					changeChannel,
				)
				state = nextState
			}

		}
	}()

	return changeChannel
}

func pushState(
	nextState string,
	prevState string,
	config *Config,
	changeChannel chan<- StateChange,
) {
	stateChange := transition(
		nextState,
		prevState,
		config,
	)
	changeChannel <- stateChange
}

func transition(
	nextState string,
	prevState string,
	config *Config,
) (
	stateChange StateChange,
) {
	for _, transitionConfigUnknown := range config.Transitions {
		switch transitionConfig := transitionConfigUnknown.(type) {
		case CommandTransitionConfig:
			if (transitionConfig.From == prevState || transitionConfig.From == "") &&
				(transitionConfig.To == nextState || transitionConfig.To == "") {
				stateChange = CommandStateChange{
					NextState: nextState,
					Command:   transitionConfig.Command,
				}
				break
			}

		case SignalTransitionConfig:
			if (transitionConfig.From == prevState || transitionConfig.From == "") &&
				(transitionConfig.To == nextState || transitionConfig.To == "") {
				stateChange = SignalStateChange{
					NextState: nextState,
					Signal:    transitionConfig.Signal,
				}
				break
			}

		case KillTransitionConfig:
			if (transitionConfig.From == prevState || transitionConfig.From == "") &&
				(transitionConfig.To == nextState || transitionConfig.To == "") {
				stateChange = KillStateChange{
					NextState: nextState,
				}
				break
			}
		}

	}

	return
}

func handleLiteralEvent(
	eventConfig *LiteralEventConfig,
	action string,
) (
	nextState string,
) {
	if eventConfig.IgnoreCase {
		if strings.ToLower(eventConfig.Value) == strings.ToLower(action) {
			nextState = eventConfig.NextState
		}
	} else {
		if eventConfig.Value == action {
			nextState = eventConfig.NextState
		}
	}
	return
}

func handleRegexEvent(
	eventConfig *RegexEventConfig,
	action string,
) (
	nextState string,
) {
	if eventConfig.Regexp.MatchString(action) {
		nextState = eventConfig.NextState
	}
	return
}

func handleTimerEvent(
	eventConfig *TimerEventConfig,
	interval time.Duration,
) (
	nextState string,
) {
	if interval > time.Duration(eventConfig.Interval) {
		nextState = eventConfig.NextState
	}
	return
}
