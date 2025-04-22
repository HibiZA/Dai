package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/HibiZA/dai/pkg/style"
	"github.com/spf13/cobra"
)

// Version contains the current version of Dai CLI
var Version = "dev"

// VersionCommand returns the version of Dai CLI
func VersionCommand() string {
	return Version
}

var rootCmd = &cobra.Command{
	Use:   "dai",
	Short: "Dai - AI‑Backed Dependency Upgrade Advisor",
	Long:  style.Banner() + "\n\nDai CLI - Automates dependency maintenance by scanning your project,\ndetecting outdated packages, and using AI to draft upgrade rationales and PRs.",
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommands are provided, show styled help
		displayColorfulHelp(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, style.Error("Error:"), err)
		os.Exit(1)
	}
}

// displayColorfulHelp shows custom styled help for the root command
func displayColorfulHelp(cmd *cobra.Command) {
	fmt.Println(style.Banner())
	fmt.Println()

	// Display description with more color
	descLine1 := style.Title("Dai CLI") + " - " + style.Info("Your AI-powered dependency management assistant")
	descLine2 := style.Subtitle("Automates dependency maintenance by scanning your project,")
	descLine3 := style.Subtitle("detecting outdated packages, and using AI to draft upgrade rationales.")

	fmt.Println(descLine1)
	fmt.Println(descLine2)
	fmt.Println(descLine3)
	fmt.Println()

	// Usage section
	fmt.Println(style.Title("Usage:"))
	fmt.Printf("  %s [flags]\n", style.Highlight("dai"))
	fmt.Printf("  %s [command]\n\n", style.Highlight("dai"))

	// Available Commands section
	fmt.Println(style.Title("Available Commands:"))

	// Get commands and find the longest name for alignment
	commands := cmd.Commands()
	longestName := 0
	for _, c := range commands {
		if !c.Hidden && len(c.Name()) > longestName {
			longestName = len(c.Name())
		}
	}

	// Add padding for better alignment
	padding := longestName + 2

	// Print each command with proper spacing
	for _, c := range commands {
		if !c.Hidden {
			name := c.Name()
			spaces := strings.Repeat(" ", padding-len(name))
			fmt.Printf("  %s%s%s\n",
				style.Package(name),
				spaces,
				style.Subtitle(c.Short))
		}
	}
	fmt.Println()

	// Flags section
	fmt.Println(style.Title("Flags:"))
	fmt.Printf("  %s    %s\n\n",
		style.Highlight("-h, --help"),
		style.Subtitle("help for dai"))

	// Footer
	fmt.Printf("Use %s for more information about a command.\n",
		style.BoldItalicize(fmt.Sprintf("\"dai [command] --help\"")))
}

// Add a version command
func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display the version of Dai CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(style.Title("Dai CLI Version:"), style.Highlight(VersionCommand()))
		},
	}

	rootCmd.AddCommand(versionCmd)
}
