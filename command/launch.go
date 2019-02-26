package command

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/elmerbulthuis/shell-go/shell"
	"github.com/elmerbulthuis/shell-go/statemachine"
	"github.com/spf13/cobra"
)

var emulateTTY bool
var configFile string

// LaunchCommand launches a process
var LaunchCommand = &cobra.Command{
	Use:   "launch",
	Short: "Start a process in the gameye-shell",
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

}

func runLaunchCommand(cmd *cobra.Command, args []string) (err error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return
	}

	proc := exec.Command(args[0], args[1:]...)

	exit, err := shell.RunWithStateMachine(
		proc,
		config,
		emulateTTY,
	)
	if err != nil {
		return
	}

	os.Exit(exit)

	return
}

func loadConfig(
	configFile string,
) (
	config *statemachine.Config,
	err error,
) {
	config = &statemachine.Config{}

	file, err := os.OpenFile(configFile, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return
	}

	return
}
