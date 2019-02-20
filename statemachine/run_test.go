package statemachine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLightStateMachine(test *testing.T) {
	config := makeLightTestConfig()

	errorChannel := make(chan error, 1)
	defer func() { assert.NoError(test, <-errorChannel) }()

	changeChannel := make(chan StateChange, 1)
	defer close(changeChannel)
	actionChannel := make(chan string, 1)
	defer close(actionChannel)

	go func() (err error) {
		defer func() { errorChannel <- err }()

		err = Run(
			config,
			changeChannel,
			actionChannel,
		)

		return
	}()

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

	errorChannel := make(chan error, 1)
	defer func() { assert.NoError(test, <-errorChannel) }()

	changeChannel := make(chan StateChange)
	defer close(changeChannel)
	actionChannel := make(chan string)
	defer close(actionChannel)

	var timer *time.Timer

	go func() (err error) {
		defer func() { errorChannel <- err }()

		err = Run(
			config,
			changeChannel,
			actionChannel,
		)

		return
	}()

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
