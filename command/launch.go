package command

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/elmerbulthuis/shell-go/shell"
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

func runLaunchCommand(
	cmd *cobra.Command,
	args []string,
) (
	err error,
) {
	config, err := loadConfig(configFile)
	if err != nil {
		return
	}

	for _, file := range config.Files {
		err = writeFile(file)
		if err != nil {
			return
		}
	}

	proc := exec.Command(config.Cmd[0], config.Cmd[1:]...)
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

func loadConfig(
	configFile string,
) (
	config *shell.Config,
	err error,
) {
	config = &shell.Config{}

	file, err := os.OpenFile(
		configFile,
		os.O_RDONLY,
		0,
	)
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
		os.O_CREATE|os.O_WRONLY,
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
