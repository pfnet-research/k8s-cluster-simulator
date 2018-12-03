package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is a compile time variable for the binary version
	Version = "N/A"
	// BuildTime is a compile time variable for the binary build time
	BuildTime = "N/A"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of kubernetes-simulator",
	Long:  "Print the version number of kubernetes-simulator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kubernetes-simulator %s (built: %s)\n", Version, BuildTime)
	},
}
