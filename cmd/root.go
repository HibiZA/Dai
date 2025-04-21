package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dai",
	Short: "Dai - AI‑Backed Dependency Upgrade Advisor",
	Long: `Dai CLI - AI‑Backed Dependency Upgrade Advisor
An open‑source CLI tool that automates dependency maintenance by scanning your project, 
detecting outdated or vulnerable packages, and using AI to draft upgrade rationales and PRs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommands are provided, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
