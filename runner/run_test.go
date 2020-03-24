package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLightRunner(test *testing.T) {
	config := makeLightTestConfig()

	actionChannel := make(chan string, 1)
	defer close(actionChannel)

	changeChannel := Run(
		config,
		actionChannel,
	)

	actionChannel <- "SwitchOn"
	assert.Equal(test, CommandStateChange{
		"On",
		"DoSwitchOn",
	}, <-changeChannel)

	actionChannel <- "SwitchOff"
	assert.Equal(test, CommandStateChange{
		"Off",
		"DoSwitchOff",
	}, <-changeChannel)

	actionChannel <- "SwitchOn"
	assert.Equal(test, CommandStateChange{
		"On",
		"DoSwitchOn",
	}, <-changeChannel)

	time.Sleep(time.Second * 2)
	assert.Equal(test, CommandStateChange{
		"Off",
		"DoSwitchOff",
	}, <-changeChannel)

}

func TestAutoRunner(test *testing.T) {
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

	assert.Equal(test, CommandStateChange{"On", "echo on"}, <-changeChannel)

	actionChannel <- "noop"
	timer = time.NewTimer(time.Second * 1)
	select {
	case <-timer.C:
	case <-changeChannel:
		assert.Fail(test, "too soon")
		return
	}

	assert.Equal(test, CommandStateChange{"Off", "echo off"}, <-changeChannel)
}
