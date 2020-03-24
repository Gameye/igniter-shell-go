package statemachine

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeStateMachineConfig(test *testing.T) {
	var err error
	defer func() {
		assert.NoError(test, err)
	}()

	file, err := os.OpenFile("light.json", os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}

	assert.NotNil(test, config)
	assert.Equal(test, *makeLightTestConfig(), config)
}

func makeLightTestConfig() (
	config *Config,
) {
	config = &Config{
		InitialState: "Off",
		States: map[string]StateConfig{
			"On": StateConfig{
				Events: []EventConfig{
					LiteralEventConfig{
						Value:     "SwitchOff",
						NextState: "Off",
					},
					TimerEventConfig{
						NextState: "Off",
						Interval:  time.Duration(time.Second * 1),
					},
				},
			},
			"Off": StateConfig{
				Events: []EventConfig{
					RegexEventConfig{
						Regexp:    regexp.MustCompile("^SwitchOn$"),
						NextState: "On",
					},
				},
			},
		},
		Transitions: []TransitionConfig{
			CommandTransitionConfig{
				From:    "Off",
				To:      "On",
				Command: "DoSwitchOn",
			},
			CommandTransitionConfig{
				From:    "On",
				To:      "Off",
				Command: "DoSwitchOff",
			},
		},
	}

	return
}

func makeAutoTestConfig() (
	config *Config,
) {
	config = &Config{
		InitialState: "Off",
		States: map[string]StateConfig{
			"On": StateConfig{
				Events: []EventConfig{
					TimerEventConfig{
						NextState: "Off",
						Interval:  time.Duration(time.Second * 2),
					},
				},
			},
			"Off": StateConfig{
				Events: []EventConfig{
					TimerEventConfig{
						NextState: "On",
						Interval:  time.Duration(time.Second * 2),
					},
				},
			},
		},
		Transitions: []TransitionConfig{
			CommandTransitionConfig{
				To:      "On",
				Command: "echo on",
			},
			CommandTransitionConfig{
				To:      "Off",
				Command: "echo off",
			},
		},
	}

	return
}
