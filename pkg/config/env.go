package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Constants for environment variables
const (
	// API Keys
	EnvOpenAIKey   = "DAI_OPENAI_API_KEY"
	EnvGitHubToken = "DAI_GITHUB_TOKEN"
	EnvNVDApiKey   = "DAI_NVD_API_KEY"

	// Configuration
	EnvLogLevel = "DAI_LOG_LEVEL"
	EnvCacheDir = "DAI_CACHE_DIR"
)

// Config holds application configuration values
type Config struct {
	// API Keys
	OpenAIApiKey string
	GitHubToken  string
	NVDApiKey    string

	// General settings
	LogLevel string
	CacheDir string
}

// LoadConfig loads configuration from environment variables and config file
func LoadConfig() *Config {
	config := &Config{
		// API Keys - try env vars first
		OpenAIApiKey: os.Getenv(EnvOpenAIKey),
		GitHubToken:  os.Getenv(EnvGitHubToken),
		NVDApiKey:    os.Getenv(EnvNVDApiKey),

		// General settings
		LogLevel: getEnvWithDefault(EnvLogLevel, "info"),
		CacheDir: getEnvWithDefault(EnvCacheDir, "./.dai-cache"),
	}

	// For backward compatibility, check non-prefixed keys if prefixed ones are empty
	if config.OpenAIApiKey == "" {
		config.OpenAIApiKey = os.Getenv("OPENAI_API_KEY")
	}
	if config.GitHubToken == "" {
		config.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}
	if config.NVDApiKey == "" {
		config.NVDApiKey = os.Getenv("NVD_API_KEY")
	}

	// If still empty, try to read from config file
	if config.OpenAIApiKey == "" || config.GitHubToken == "" || config.NVDApiKey == "" {
		readConfigFromFile(config)
	}

	return config
}

// readConfigFromFile reads configuration from the config.env file
func readConfigFromFile(config *Config) {
	// Get config file path
	configFile := getConfigFilePath()

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return
	}

	// Read file
	content, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	// Parse file
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set values in config if not already set from environment
		switch key {
		case EnvOpenAIKey:
			if config.OpenAIApiKey == "" {
				config.OpenAIApiKey = value
			}
		case EnvGitHubToken:
			if config.GitHubToken == "" {
				config.GitHubToken = value
			}
		case EnvNVDApiKey:
			if config.NVDApiKey == "" {
				config.NVDApiKey = value
			}
		case EnvLogLevel:
			if config.LogLevel == "info" { // Only override if default
				config.LogLevel = value
			}
		case EnvCacheDir:
			if config.CacheDir == "./.dai-cache" { // Only override if default
				config.CacheDir = value
			}
		}
	}
}

// getConfigFilePath returns the path to the config.env file
func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".dai/config.env"
	}

	// Check for XDG_CONFIG_HOME environment variable first (Linux/macOS standard)
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "dai", "config.env")
	}

	// On macOS and Linux, use ~/.config/dai
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, ".config", "dai", "config.env")
	}

	// On Windows, use %APPDATA%\dai
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Roaming", "dai", "config.env")
	}

	// Fallback to the original .dai directory for backward compatibility
	return filepath.Join(homeDir, ".dai", "config.env")
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// HasOpenAIKey returns true if the OpenAI API key is set
func (c *Config) HasOpenAIKey() bool {
	return c.OpenAIApiKey != ""
}

// HasGitHubToken returns true if the GitHub token is set
func (c *Config) HasGitHubToken() bool {
	return c.GitHubToken != ""
}

// HasNVDApiKey returns true if the NVD API key is set
func (c *Config) HasNVDApiKey() bool {
	return c.NVDApiKey != ""
}
