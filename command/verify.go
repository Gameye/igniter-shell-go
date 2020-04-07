package command

import (
	"github.com/spf13/cobra"
)

// VerifyCommand verifies a config file
var VerifyCommand = &cobra.Command{
	Use:   "verify",
	Short: "Verify a config-file",
	RunE:  runVerifyCommand,
}

func init() {
	RootCommand.AddCommand(VerifyCommand)

	VerifyCommand.
		PersistentFlags().
		StringVarP(
			&configFile,
			"config-file",
			"c",
			"",
			"Path to config file",
		)
}

func runVerifyCommand(
	cmd *cobra.Command,
	args []string,
) (
	err error,
) {
	_, err = loadConfig(
		configFile,
	)
	if err != nil {
		return
	}

	println("seems to be ok :-)")

	return
}
