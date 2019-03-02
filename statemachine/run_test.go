package statemachine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLightStateMachine(test *testing.T) {
	config := makeLightTestConfig()

	actionChannel := make(chan string, 1)
	defer close(actionChannel)

	changeChannel := Run(
		config,
		actionChannel,
	)

	actionChannel <- "SwitchOn"
	assert.Equal(test, StateChange{
		"On",
		"DoSwitchOn",
	}, <-changeChannel)

	actionChannel <- "SwitchOff"
	assert.Equal(test, StateChange{
		"Off",
		"DoSwitchOff",
	}, <-changeChannel)

	actionChannel <- "SwitchOn"
	assert.Equal(test, StateChange{
		"On",
		"DoSwitchOn",
	}, <-changeChannel)

	time.Sleep(time.Second * 2)
	assert.Equal(test, StateChange{
		"Off",
		"DoSwitchOff",
	}, <-changeChannel)

}

func TestAutoStateMachine(test *testing.T) {
	config := makeAutoTestConfig()

	actionChannel := make(chan string)
	defer close(actionChannel)

	var timer *time.Timer

	changeChannel := Run(
		config,
		actionChannel,
	)

	actionChannel <- "noop"
	timer = time.NewTimer(time.Second * 1)
	select {
	case <-timer.C:
	case <-changeChannel:
		assert.Fail(test, "too soon")
		return
	}

	assert.Equal(test, StateChange{"On", "echo on"}, <-changeChannel)

	actionChannel <- "noop"
	timer = time.NewTimer(time.Second * 1)
	select {
	case <-timer.C:
	case <-changeChannel:
		assert.Fail(test, "too soon")
		return
	}

	assert.Equal(test, StateChange{"Off", "echo off"}, <-changeChannel)
}
