package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/elmerbulthuis/shell-go/shell"
	"github.com/elmerbulthuis/shell-go/statemachine"
)

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

	exit, err = shell.RunWithStateMachine(cmd, config, true)
	if err != nil {
		return
	}

	return
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
