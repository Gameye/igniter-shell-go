package statemachine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStateMachineConfig(test *testing.T) {
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
