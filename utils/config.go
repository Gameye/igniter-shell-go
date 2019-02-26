package utils

import (
	"time"

	"github.com/elmerbulthuis/shell-go/statemachine"
)

// MakeTestConfig for testing purpose
func MakeTestConfig() (
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

// MakeTf2Config for testing purpose
func MakeTf2Config() (
	config *statemachine.Config,
) {
	config = &statemachine.Config{
		InitialState: "idle",
		States: map[string]statemachine.StateConfig{
			"idle": statemachine.StateConfig{
				Events: []statemachine.EventStateConfig{
					statemachine.LiteralEventConfig{
						NextState:  "end",
						Value:      "'server.cfg' not present; not executing.",
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
				Command: "echo 'Quit in 10 seconds'",
			},
			statemachine.TransitionConfig{
				To:      "quit",
				Command: "quit",
			},
		},
	}

	return
}
