package command

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/Gameye/igniter-shell-go/utils"

	"github.com/Gameye/igniter-shell-go/shell"
	"github.com/spf13/cobra"
)

var emulateTTY bool
var configFile string
var variableList *[]string

// LaunchCommand launches a process
var LaunchCommand = &cobra.Command{
	Use:   "launch",
	Short: "Start a process in the igniter-shell",
	RunE:  runLaunchCommand,
}

func init() {
	RootCommand.AddCommand(LaunchCommand)

	LaunchCommand.
		PersistentFlags().
		BoolVarP(
			&emulateTTY,
			"emulate-tty",
			"t",
			false,
			"Emulate a TTY for the child process",
		)

	LaunchCommand.
		PersistentFlags().
		StringVarP(
			&configFile,
			"config-file",
			"c",
			"",
			"Path to config file",
		)

	variableList = LaunchCommand.
		PersistentFlags().
		StringArrayP(
			"variable",
			"v",
			[]string{},
			"The variables which should be replaced in the files and the extra process arguments specified in the config. \nCan be passed multiple times for multiple variables. \nEach variable should have the format key=value",
		)

}

func runLaunchCommand(
	cmd *cobra.Command,
	args []string,
) (
	err error,
) {
	config, err := loadConfig(
		configFile,
	)
	if err != nil {
		return
	}

	variables := make(map[string]string)
	for key, value := range config.Defaults {
		variables[key] = value
	}
	for _, variableItem := range *variableList {
		pair := strings.SplitN(variableItem, "=", 2)
		variables[pair[0]] = pair[1]
	}

	renderConfigTemplate(
		config,
		variables,
	)

	for _, file := range config.Files {
		err = writeFile(file)
		if err != nil {
			return
		}
	}

	commandArgs := append(args, config.Cmd...)
	proc := exec.Command(
		commandArgs[0],
		commandArgs[1:]...,
	)
	proc.Env = os.Environ()

	exit, err := shell.RunWithStateMachine(
		proc,
		config.Script,
		emulateTTY,
	)
	if err != nil {
		return
	}

	os.Exit(exit)

	return
}

func renderConfigTemplate(
	config *shell.Config,
	variables map[string]string,
) {
	for index := range config.Cmd {
		config.Cmd[index] = utils.RenderTemplate(
			config.Cmd[index],
			variables,
		)
	}

	for index := range config.Files {
		config.Files[index].Content = utils.RenderTemplate(
			config.Files[index].Content,
			variables,
		)
	}

	for index := range config.Script.Transitions {
		config.Script.Transitions[index].Command = utils.RenderTemplate(
			config.Script.Transitions[index].Command,
			variables,
		)
	}
}

func loadConfig(
	configFile string,
) (
	config *shell.Config,
	err error,
) {
	config = &shell.Config{}
	yamlData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	jsonData, err := yaml.YAMLToJSON(yamlData)

	err = json.Unmarshal(jsonData, config)
	if err != nil {
		return
	}

	return
}

func writeFile(
	fileConfig shell.FileConfig,
) (
	err error,
) {
	err = os.MkdirAll(filepath.Dir(fileConfig.Path), 0755)
	if err != nil {
		return
	}

	file, err := os.OpenFile(
		fileConfig.Path,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_SYNC,
		0755,
	)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.WriteString(fileConfig.Content)
	if err != nil {
		return
	}

	return
}
