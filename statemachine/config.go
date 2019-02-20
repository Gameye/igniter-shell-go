package statemachine

import (
	"encoding/json"
	"time"
)

/*
EventStateConfig is the configuration for an event
*/
type EventStateConfig interface{}

/*
Config is the configuration for a state machine
*/
type Config struct {
	InitialState string                 `json:"initialState"`
	States       map[string]StateConfig `json:"states"`
	Transitions  []TransitionConfig     `json:"transitions"`
}

/*
TransitionConfig is a transition from one state to another.
*/
type TransitionConfig struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Command string `json:"command"`
}

/*
StateConfig is a transition from one state to another.
*/
type StateConfig struct {
	Events []EventStateConfig `json:"events"`
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *StateConfig) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var root interface{}
	err = json.Unmarshal(data, &root)
	if err != nil {
		return
	}

	err = config.parse(root)
	return
}

/*
LiteralEventConfig configures literal events
*/
type LiteralEventConfig struct {
	NextState  string `json:"nextState"`
	Value      string `json:"value"`
	IgnoreCase bool   `json:"ignoreCase"`
}

/*
RegexEventConfig configures regex events
*/
type RegexEventConfig struct {
	NextState  string `json:"nextState"`
	Pattern    string `json:"pattern"`
	IgnoreCase bool   `json:"ignoreCase"`
}

/*
TimerEventConfig configures timer events
*/
type TimerEventConfig struct {
	NextState string        `json:"nextState"`
	Interval  time.Duration `json:"interval"`
}
