package ai

import (
	"fmt"
	"strings"

	"github.com/HibiZA/dai/pkg/config"
)

// DebugOpenAIKey checks the OpenAI API key and returns debug information
func DebugOpenAIKey() string {
	cfg := config.LoadConfig()
	var result strings.Builder

	result.WriteString(fmt.Sprintf("HasOpenAIKey: %v\n", cfg.HasOpenAIKey()))

	// Check where the key is coming from
	envValue := cfg.OpenAIApiKey
	if envValue != "" {
		// Safely mask the key
		masked := "****"
		if len(envValue) > 8 {
			masked = envValue[:4] + "..." + envValue[len(envValue)-4:]
		}
		result.WriteString(fmt.Sprintf("OpenAI Key present: %s\n", masked))
	} else {
		result.WriteString("OpenAI Key not present\n")
	}

	// Check env vars directly
	daiKey := "Not set"
	if val := cfg.OpenAIApiKey; val != "" {
		daiKey = "Set (length: " + fmt.Sprintf("%d", len(val)) + ")"
	}
	result.WriteString(fmt.Sprintf("DAI_OPENAI_API_KEY: %s\n", daiKey))

	return result.String()
}
