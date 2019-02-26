package command

import (
	"github.com/elmerbulthuis/shell-go/resource"
	"github.com/spf13/cobra"
)

// RootCommand is the root command for all subcommands
var RootCommand = &cobra.Command{
	Use:     "gameye-shell",
	Short:   "The gameye shell starts and configures game servers.",
	Version: resource.Version,
}
