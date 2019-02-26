package command

import (
	"os"
	"os/exec"

	"github.com/elmerbulthuis/shell-go/shell"
	"github.com/elmerbulthuis/shell-go/utils"
	"github.com/spf13/cobra"
)

var emulateTTY bool

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
}

func runLaunchCommand(cmd *cobra.Command, args []string) (err error) {
	proc := exec.Command(args[0], args[1:]...)
	config := utils.MakeTf2Config()

	exit, err := shell.RunWithStateMachine(proc, config, emulateTTY)
	if err != nil {
		return
	}

	os.Exit(exit)

	return
}
