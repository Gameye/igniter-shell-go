package statemachine

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

/*
ErrMissingStateConfig occurs when a state is referenced that does not
exist in the config
*/
var ErrMissingStateConfig = errors.New("missing state config")

/*
StateChange contains information on a changes state
*/
type StateChange struct {
	NextState string
	Command   string
}

/*
Run runs a new StateMachine
*/
func Run(
	config *Config,
	changeChannel chan<- StateChange,
	actionChannel <-chan string,
) (
	err error,
) {
	state := config.InitialState
	for {
		start := time.Now()

		// find state config
		stateConfig, hasStateConfig := config.States[state]
		if !hasStateConfig {
			err = ErrMissingStateConfig
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
						nextState, err = handleRegexEvent(&eventConfig, action)
						if err != nil {
							return
						}
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
}

func pushState(
	nextState string,
	prevState string,
	config *Config,
	changeChannel chan<- StateChange,
) {
	command := transition(
		nextState,
		prevState,
		config,
	)
	changeChannel <- StateChange{
		nextState,
		command,
	}
}

func transition(
	nextState string,
	prevState string,
	config *Config,
) (
	command string,
) {
	for _, transition := range config.Transitions {
		if (transition.From == prevState || transition.From == "") &&
			(transition.To == nextState || transition.To == "") {
			command = transition.Command
			break
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
		if eventConfig.IgnoreCase && strings.ToLower(eventConfig.Value) == strings.ToLower(action) {
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
	err error,
) {
	var re *regexp.Regexp
	if eventConfig.IgnoreCase {
		re, err = regexp.Compile(strings.ToLower(eventConfig.Pattern))
		if err != nil {
			return
		}
		if re.MatchString(strings.ToLower(action)) {
			nextState = eventConfig.NextState
		}
	} else {
		re, err = regexp.Compile(eventConfig.Pattern)
		if err != nil {
			return
		}
		if re.MatchString(action) {
			nextState = eventConfig.NextState
		}
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
