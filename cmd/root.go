package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile microservice",
	Long:  "A profile microservice for managing user profiles via HTTP and gRPC.",
}

// Execute runs the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
