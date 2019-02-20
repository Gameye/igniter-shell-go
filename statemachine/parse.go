package statemachine

import (
	"errors"
	"time"
)

/*
ErrConfigParse occurs when a config could not be parsed
*/
var ErrConfigParse = errors.New("parse error")

func (config *StateConfig) parse(
	root interface{},
) (
	err error,
) {
	switch typedRoot := root.(type) {
	case map[string]interface{}:
		for key, value := range typedRoot {
			switch key {

			case "events":
				switch typedValue := value.(type) {
				case []interface{}:
					config.Events = make(
						[]EventStateConfig,
						len(typedValue),
					)
					for subKey, subValue := range typedValue {
						switch typedValue := subValue.(type) {
						case map[string]interface{}:
							switch typedValue["type"] {
							case "literal":
								var subConfig LiteralEventConfig
								err = subConfig.parse(subValue)
								config.Events[subKey] = subConfig
								if err != nil {
									return
								}

							case "regex":
								var subConfig RegexEventConfig
								err = subConfig.parse(subValue)
								config.Events[subKey] = subConfig
								if err != nil {
									return
								}

							case "timer":
								var subConfig TimerEventConfig
								err = subConfig.parse(subValue)
								config.Events[subKey] = subConfig
								if err != nil {
									return
								}

							default:
								err = ErrConfigParse
								return

							}
						default:
							err = ErrConfigParse
							return
						}
					}
				default:
					err = ErrConfigParse
					return
				}

			}
		}

	default:
		err = ErrConfigParse
		return
	}

	return
}

func (config *LiteralEventConfig) parse(
	root interface{},
) (
	err error,
) {
	switch typedRoot := root.(type) {
	case map[string]interface{}:
		for key, value := range typedRoot {
			switch key {
			case "type":
			case "value":
				switch typedValue := value.(type) {
				case string:
					config.Value = typedValue
				default:
					err = ErrConfigParse
					return
				}
			case "ignoreCase":
				switch typedValue := value.(type) {
				case bool:
					config.IgnoreCase = typedValue
				default:
					err = ErrConfigParse
					return
				}
			case "nextState":
				switch typedValue := value.(type) {
				case string:
					config.NextState = typedValue
				default:
					err = ErrConfigParse
					return
				}

			}

		}

	default:
		err = ErrConfigParse
		return
	}

	return
}

func (config *RegexEventConfig) parse(
	root interface{},
) (
	err error,
) {
	switch typedRoot := root.(type) {
	case map[string]interface{}:
		for key, value := range typedRoot {
			switch key {
			case "type":
			case "pattern":
				switch typedValue := value.(type) {
				case string:
					config.Pattern = typedValue
				default:
					err = ErrConfigParse
					return
				}
			case "ignoreCase":
				switch typedValue := value.(type) {
				case bool:
					config.IgnoreCase = typedValue
				default:
					err = ErrConfigParse
					return
				}
			case "nextState":
				switch typedValue := value.(type) {
				case string:
					config.NextState = typedValue
				default:
					err = ErrConfigParse
					return
				}

			}

		}

	default:
		err = ErrConfigParse
		return
	}

	return
}

func (config *TimerEventConfig) parse(
	root interface{},
) (
	err error,
) {
	switch typedRoot := root.(type) {
	case map[string]interface{}:
		for key, value := range typedRoot {
			switch key {
			case "type":
			case "interval":
				switch typedValue := value.(type) {
				case float64:
					if err != nil {
						return err
					}
					config.Interval = time.Duration(float64(time.Millisecond) * typedValue)
				default:
					err = ErrConfigParse
					return
				}
			case "nextState":
				switch typedValue := value.(type) {
				case string:
					config.NextState = typedValue
				default:
					err = ErrConfigParse
					return
				}

			}

		}

	default:
		err = ErrConfigParse
		return
	}

	return
}
