package config

import (
	"os"
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

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		// API Keys
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

	return config
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
