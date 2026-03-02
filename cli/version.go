package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the mxreq version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mxreq %s\n", Version)
	},
}
