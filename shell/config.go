package shell

import (
	"github.com/elmerbulthuis/shell-go/statemachine"
)

/*
Config is a configuration
*/
type Config struct {
	Defaults map[string]string    `json:"defaults"`
	Cmd      []string             `json:"cmd"`
	Files    []FileConfig         `json:"files"`
	Script   *statemachine.Config `json:"script"`
}

/*
FileConfig file configuration
*/
type FileConfig struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
