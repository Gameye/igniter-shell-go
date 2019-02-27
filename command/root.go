package command

import (
	"github.com/Gameye/igniter-shell-go/resource"
	"github.com/spf13/cobra"
)

// RootCommand is the root command for all subcommands
var RootCommand = &cobra.Command{
	Use:     "igniter-shell",
	Short:   "The Gameye igniter shell starts and configures game servers.",
	Version: resource.Version,
}
