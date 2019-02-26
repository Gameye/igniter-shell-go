package command

import (
	"os"
	"os/exec"

	"github.com/elmerbulthuis/shell-go/utils"
	"github.com/spf13/cobra"
)

// LaunchCommand launches a process
var LaunchCommand = &cobra.Command{
	Use:   "launch",
	Short: "Start a process in the gameye-shell",
	RunE:  runLaunchCommand,
}

func init() {
	RootCommand.AddCommand(LaunchCommand)
}

func runLaunchCommand(cmd *cobra.Command, args []string) (err error) {
	proc := exec.Command("docker", "run", "-ti", "docker.gameye.com/tf2")
	config := utils.MakeTf2Config()

	exit, err := utils.RunWithStateMachine(proc, config)
	if err != nil {
		return
	}

	os.Exit(exit)

	return
}
