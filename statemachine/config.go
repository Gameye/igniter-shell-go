package statemachine

import (
	"encoding/json"
	"time"
)

/*
Duration is a duration
*/
type Duration time.Duration

/*
UnmarshalJSON provides custom unmarshalling
*/
func (target *Duration) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var source float64

	err = json.Unmarshal(data, &source)
	if err != nil {
		return
	}

	*target = Duration(time.Duration(float64(time.Millisecond) * source))

	return
}

/*
Config is the configuration for a state machine
*/
type Config struct {
	InitialState string               `json:"initialState"`
	States       StateConfigMap       `json:"states"`
	Transitions  TransitionConfigList `json:"transitions"`
}

/*
TransitionConfigList list of TransitionConfig
*/
type TransitionConfigList []TransitionConfig

/*
TransitionConfig is a transition from one state to another.
*/
type TransitionConfig struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Command string `json:"command"`
}

/*
StateConfigMap map of StateConfig
*/
type StateConfigMap map[string]StateConfig

/*
StateConfig is a transition from one state to another.
*/
type StateConfig struct {
	Events EventStateConfigList `json:"events"`
}

/*
EventStateConfigList list of EventStateConfig
*/
type EventStateConfigList []EventStateConfig

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *EventStateConfigList) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var list []EventStateConfigJSON
	err = json.Unmarshal(data, &list)
	if err != nil {
		return
	}

	for _, item := range list {
		*config = append(*config, item.Payload)
	}

	return
}

/*
EventStateConfig is the configuration for an event
*/
type EventStateConfig interface{}

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
	NextState string   `json:"nextState"`
	Interval  Duration `json:"interval"`
}

/*
EventStateConfigJSON helper
*/
type EventStateConfigJSON struct {
	Payload EventStateConfig
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *EventStateConfigJSON) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var item struct {
		Type string `json:"type"`
	}

	err = json.Unmarshal(data, &item)
	if err != nil {
		return
	}

	switch item.Type {
	case "literal":
		var payload LiteralEventConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload
	case "regex":
		var payload RegexEventConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload
	case "timer":
		var payload TimerEventConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload
	}

	return
}
