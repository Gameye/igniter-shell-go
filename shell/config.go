package shell

import (
	"github.com/elmerbulthuis/shell-go/statemachine"
)

/*
Config is a configuration
*/
type Config struct {
	Script *statemachine.Config `json:"script"`
}
