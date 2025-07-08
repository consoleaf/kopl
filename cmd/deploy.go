package cmd

import (
	internalerror "github.com/Consoleaf/kopl/internal_error"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy project to a device",
	Run: func(cmd *cobra.Command, args []string) {
		internalerror.ErrorExit("Not implemented yet")
	},
}
