package shell

import (
	"github.com/Gameye/igniter-shell-go/runner"
)

/*
Config is a configuration
*/
type Config struct {
	Defaults map[string]string `json:"defaults"`
	Cmd      []string          `json:"cmd"`
	Env      map[string]string `json:"env"`
	Files    []FileConfig      `json:"files"`
	Script   *runner.Config    `json:"script"`
}

/*
FileConfig file configuration
*/
type FileConfig struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
