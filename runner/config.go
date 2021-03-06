package runner

import (
	"encoding/json"
	"os"
	"regexp"
	"syscall"
	"time"
)

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
UnmarshalJSON provides custom unmarshalling
*/
func (config *TransitionConfigList) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var list []transitionConfigJSON
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
TransitionConfig is a transition from one state to another.
*/
type TransitionConfig interface{}

/*
CommandTransitionConfig transitions with a command
*/
type CommandTransitionConfig struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Command string `json:"command"`
}

/*
KillTransitionConfig transitions by killing the process
*/
type KillTransitionConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

/*
SignalTransitionConfig transitions with a signal
*/
type SignalTransitionConfig struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Signal os.Signal `json:"signal"`
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (target *SignalTransitionConfig) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var source struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Signal string `json:"signal"`
	}

	err = json.Unmarshal(data, &source)
	if err != nil {
		return
	}

	var signal os.Signal
	switch source.Signal {

	case "SIGABRT":
		signal = syscall.Signal(0x6)
	case "SIGALRM":
		signal = syscall.Signal(0xe)
	case "SIGBUS":
		signal = syscall.Signal(0x7)
	case "SIGCHLD":
		signal = syscall.Signal(0x11)
	case "SIGCLD":
		signal = syscall.Signal(0x11)
	case "SIGCONT":
		signal = syscall.Signal(0x12)
	case "SIGFPE":
		signal = syscall.Signal(0x8)
	case "SIGHUP":
		signal = syscall.Signal(0x1)
	case "SIGILL":
		signal = syscall.Signal(0x4)
	case "SIGINT":
		signal = syscall.Signal(0x2)
	case "SIGIO":
		signal = syscall.Signal(0x1d)
	case "SIGIOT":
		signal = syscall.Signal(0x6)
	case "SIGKILL":
		signal = syscall.Signal(0x9)
	case "SIGPIPE":
		signal = syscall.Signal(0xd)
	case "SIGPOLL":
		signal = syscall.Signal(0x1d)
	case "SIGPROF":
		signal = syscall.Signal(0x1b)
	case "SIGPWR":
		signal = syscall.Signal(0x1e)
	case "SIGQUIT":
		signal = syscall.Signal(0x3)
	case "SIGSEGV":
		signal = syscall.Signal(0xb)
	case "SIGSTKFLT":
		signal = syscall.Signal(0x10)
	case "SIGSTOP":
		signal = syscall.Signal(0x13)
	case "SIGSYS":
		signal = syscall.Signal(0x1f)
	case "SIGTERM":
		signal = syscall.Signal(0xf)
	case "SIGTRAP":
		signal = syscall.Signal(0x5)
	case "SIGTSTP":
		signal = syscall.Signal(0x14)
	case "SIGTTIN":
		signal = syscall.Signal(0x15)
	case "SIGTTOU":
		signal = syscall.Signal(0x16)
	case "SIGUNUSED":
		signal = syscall.Signal(0x1f)
	case "SIGURG":
		signal = syscall.Signal(0x17)
	case "SIGUSR1":
		signal = syscall.Signal(0xa)
	case "SIGUSR2":
		signal = syscall.Signal(0xc)
	case "SIGVTALRM":
		signal = syscall.Signal(0x1a)
	case "SIGWINCH":
		signal = syscall.Signal(0x1c)
	case "SIGXCPU":
		signal = syscall.Signal(0x18)
	case "SIGXFSZ":
		signal = syscall.Signal(0x19)
	}

	*target = SignalTransitionConfig{
		From:   source.From,
		To:     source.To,
		Signal: signal,
	}

	return
}

/*
transitionConfigJSON helper
*/
type transitionConfigJSON struct {
	Payload EventConfig
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *transitionConfigJSON) UnmarshalJSON(
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

	case "command":
		var payload CommandTransitionConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload

	case "signal":
		var payload SignalTransitionConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload

	case "kill":
		var payload KillTransitionConfig
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return
		}
		config.Payload = payload

	}

	return
}

/*
StateConfigMap map of StateConfig
*/
type StateConfigMap map[string]StateConfig

/*
StateConfig is a transition from one state to another.
*/
type StateConfig struct {
	Events EventConfigList `json:"events"`
}

/*
EventConfigList list of EventConfig
*/
type EventConfigList []EventConfig

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *EventConfigList) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var list []eventConfigJSON
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
EventConfig is the configuration for an event
*/
type EventConfig interface{}

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
	NextState string
	Regexp    *regexp.Regexp
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (target *RegexEventConfig) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var source struct {
		NextState  string `json:"nextState"`
		Pattern    string `json:"pattern"`
		IgnoreCase bool   `json:"ignoreCase"`
	}

	err = json.Unmarshal(data, &source)
	if err != nil {
		return
	}

	var re *regexp.Regexp
	if source.IgnoreCase {
		re, err = regexp.Compile("(?i)" + source.Pattern)
		if err != nil {
			return
		}
	} else {
		re, err = regexp.Compile(source.Pattern)
		if err != nil {
			return
		}
	}
	*target = RegexEventConfig{
		NextState: source.NextState,
		Regexp:    re,
	}

	return
}

/*
TimerEventConfig configures timer events
*/
type TimerEventConfig struct {
	NextState string
	Interval  time.Duration
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (target *TimerEventConfig) UnmarshalJSON(
	data []byte,
) (
	err error,
) {
	var source struct {
		NextState string  `json:"nextState"`
		Interval  float64 `json:"interval"`
	}

	err = json.Unmarshal(data, &source)
	if err != nil {
		return
	}

	*target = TimerEventConfig{
		NextState: source.NextState,
		Interval:  time.Duration(float64(time.Millisecond) * source.Interval),
	}

	return
}

/*
eventConfigJSON helper
*/
type eventConfigJSON struct {
	Payload EventConfig
}

/*
UnmarshalJSON provides custom unmarshalling
*/
func (config *eventConfigJSON) UnmarshalJSON(
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
