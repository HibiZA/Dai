package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/HibiZA/dai/pkg/config"
	"github.com/HibiZA/dai/pkg/style"
	"github.com/spf13/cobra"
)

var (
	setKey       string
	listConfig   bool
	openaiKeyArg string
	githubKeyArg string
)

func init() {
	// Add config flags
	configCmd.Flags().StringVarP(&setKey, "set", "s", "", "Set a config key (openai, github)")
	configCmd.Flags().BoolVarP(&listConfig, "list", "l", false, "List current configuration")
	configCmd.Flags().StringVar(&openaiKeyArg, "openai-key", "", "OpenAI API key to set")
	configCmd.Flags().StringVar(&githubKeyArg, "github-token", "", "GitHub token to set")

	// Set custom help function
	configCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		displayConfigHelp(cmd)
	})

	rootCmd.AddCommand(configCmd)
}

// displayConfigHelp shows custom styled help for the config command
func displayConfigHelp(cmd *cobra.Command) {
	fmt.Println(style.Banner())
	fmt.Println()

	// Command name
	fmt.Printf("%s - %s\n\n", style.Title("Config Command"), style.Subtitle("Configuration Management"))

	// Description
	fmt.Println(style.Info("Set and manage Dai CLI configuration options including API keys for OpenAI and GitHub."))
	fmt.Println()

	// Usage section
	fmt.Println(style.Title("Usage:"))
	fmt.Printf("  %s\n\n", style.Highlight("dai config [flags]"))

	// Flags section
	fmt.Println(style.Title("Flags:"))
	fmt.Printf("  %-30s %s\n",
		style.Highlight("--github-token string"),
		style.Subtitle("GitHub token to set"))
	fmt.Printf("  %-30s %s\n",
		style.Highlight("-h, --help"),
		style.Subtitle("Help for config command"))
	fmt.Printf("  %-30s %s\n",
		style.Highlight("-l, --list"),
		style.Subtitle("List current configuration"))
	fmt.Printf("  %-30s %s\n",
		style.Highlight("--openai-key string"),
		style.Subtitle("OpenAI API key to set"))
	fmt.Printf("  %-30s %s\n\n",
		style.Highlight("-s, --set string"),
		style.Subtitle("Set a config key (openai, github)"))

	// Examples section
	fmt.Println(style.Title("Examples:"))
	fmt.Printf("  %s\n", style.Subtitle("# Set OpenAI API key"))
	fmt.Printf("  %s\n\n", style.Highlight("dai config --set openai --openai-key YOUR_API_KEY"))

	fmt.Printf("  %s\n", style.Subtitle("# Set GitHub token"))
	fmt.Printf("  %s\n\n", style.Highlight("dai config --set github --github-token YOUR_GITHUB_TOKEN"))

	fmt.Printf("  %s\n", style.Subtitle("# List current configuration"))
	fmt.Printf("  %s\n", style.Highlight("dai config --list"))
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure API keys and other settings",
	Long:  `Configure Dai CLI options including API keys for OpenAI and GitHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		if listConfig {
			displayConfiguration()
			return
		}

		if setKey != "" {
			switch strings.ToLower(setKey) {
			case "openai":
				if openaiKeyArg == "" {
					fmt.Println(style.Error("Error:"), "OpenAI API key is required. Use --openai-key flag.")
					return
				}
				saveAPIKey("openai", openaiKeyArg)

			case "github":
				if githubKeyArg == "" {
					fmt.Println(style.Error("Error:"), "GitHub token is required. Use --github-token flag.")
					return
				}
				saveAPIKey("github", githubKeyArg)

			default:
				fmt.Println(style.Error("Error:"), "Unknown key. Supported keys: openai, github")
			}
			return
		}

		// If no flags provided, show help
		cmd.Help()
	},
}

// displayConfiguration shows the current configuration
func displayConfiguration() {
	fmt.Println(style.Title("Current Configuration:"))
	fmt.Println(style.Divider())

	// Get configuration from environment variables
	cfg := config.LoadConfig()

	// Display OpenAI API key status
	if cfg.HasOpenAIKey() {
		fmt.Printf("%s: %s\n",
			style.Package("OpenAI API Key"),
			style.Success("Set ✓"))
	} else {
		fmt.Printf("%s: %s\n",
			style.Package("OpenAI API Key"),
			style.Warning("Not set"))
	}

	// Display GitHub token status
	if cfg.HasGitHubToken() {
		fmt.Printf("%s: %s\n",
			style.Package("GitHub Token"),
			style.Success("Set ✓"))
	} else {
		fmt.Printf("%s: %s\n",
			style.Package("GitHub Token"),
			style.Warning("Not set"))
	}

	// Display environment variables
	fmt.Println()
	fmt.Println(style.Subtitle("Environment Variables:"))
	fmt.Printf("  %s=%s\n", style.Info("DAI_OPENAI_API_KEY"), maskSecret(os.Getenv("DAI_OPENAI_API_KEY")))
	fmt.Printf("  %s=%s\n", style.Info("DAI_GITHUB_TOKEN"), maskSecret(os.Getenv("DAI_GITHUB_TOKEN")))

	// Display config file information
	configDir := getConfigDir()
	configFile := filepath.Join(configDir, "config.env")

	fmt.Println()
	fmt.Println(style.Subtitle("Config File:"))
	fmt.Printf("  %s\n", style.Info(configFile))
}

// maskSecret masks a secret for display, showing only first and last few characters
func maskSecret(secret string) string {
	if secret == "" {
		return "(not set)"
	}

	if len(secret) <= 8 {
		return "****"
	}

	// Show first 3 and last 3 characters
	return secret[:3] + "..." + secret[len(secret)-3:]
}

// SaveAPIKey saves an API key to the config file (exported version for use by other commands)
func SaveAPIKey(keyType, value string) error {
	// Create config directory if it doesn't exist
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Determine environment variable name
	var envName string
	switch keyType {
	case "openai":
		envName = "DAI_OPENAI_API_KEY"
	case "github":
		envName = "DAI_GITHUB_TOKEN"
	default:
		return fmt.Errorf("unknown key type: %s", keyType)
	}

	// Save to config file
	configFile := filepath.Join(configDir, "config.env")

	// Load existing config if it exists
	existingConfig := make(map[string]string)
	if content, err := os.ReadFile(configFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existingConfig[parts[0]] = parts[1]
			}
		}
	}

	// Update config
	existingConfig[envName] = value

	// Write back to file
	var configContent strings.Builder
	configContent.WriteString("# Dai CLI Configuration\n")
	configContent.WriteString("# Generated automatically - DO NOT EDIT\n\n")

	for k, v := range existingConfig {
		configContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	if err := os.WriteFile(configFile, []byte(configContent.String()), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Also set for current session
	os.Setenv(envName, value)

	return nil
}

// saveAPIKey saves an API key to the config file (internal version used by config command)
func saveAPIKey(keyType, value string) {
	err := SaveAPIKey(keyType, value)
	if err != nil {
		fmt.Println(style.Error("Error:"), err)
		return
	}

	// Determine environment variable name for display purposes
	var envName string
	switch keyType {
	case "openai":
		envName = "DAI_OPENAI_API_KEY"
	case "github":
		envName = "DAI_GITHUB_TOKEN"
	}

	// Get the config file path for display
	configDir := getConfigDir()
	configFile := filepath.Join(configDir, "config.env")

	fmt.Printf("%s %s %s\n",
		style.Success("✓"),
		style.Title(keyType+" key saved to"),
		style.Info(configFile))
	fmt.Println(style.Subtitle("To use in all terminal sessions, add to your shell profile:"))
	fmt.Printf("  %s\n", style.Highlight(fmt.Sprintf("export %s=%s", envName, value)))
}

// getConfigDir returns the path to the config directory
func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir can't be determined
		return ".dai"
	}

	// Check for XDG_CONFIG_HOME environment variable first (Linux/macOS standard)
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "dai")
	}

	// On macOS and Linux, use ~/.config/dai
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, ".config", "dai")
	}

	// On Windows, use %APPDATA%\dai
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Roaming", "dai")
	}

	// Fallback to the original .dai directory for backward compatibility
	return filepath.Join(homeDir, ".dai")
}
