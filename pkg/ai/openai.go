package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/HibiZA/dai/pkg/config"
	"github.com/HibiZA/dai/pkg/semver"
)

// Upgrade represents a package version upgrade
type Upgrade struct {
	Package     string
	FromVersion string
	ToVersion   string
	Rationale   string
}

// AiClient defines the interface for AI operations
type AiClient interface {
	GenerateUpgradeRationale(packageName, oldVersion, newVersion string, changes []string) (string, error)
	GeneratePRDescription(upgrades map[string]VersionUpgrade) (string, error)
}

// VersionUpgrade represents a version upgrade for a package
type VersionUpgrade struct {
	PackageName string
	OldVersion  string
	NewVersion  string
	Rationale   string
	Breaking    bool
}

// OpenAiClient is a client for the OpenAI API
type OpenAiClient struct {
	Client *openai.Client
	Model  string
}

// NewOpenAiClient creates a new OpenAI client
func NewOpenAiClient(cfg *config.Config) (*OpenAiClient, error) {
	if !cfg.HasOpenAIKey() {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	client := openai.NewClient(cfg.OpenAIApiKey)

	return &OpenAiClient{
		Client: client,
		Model:  "gpt-3.5-turbo", // Default model
	}, nil
}

// GenerateUpgradeRationale generates a rationale for upgrading a package
func (c *OpenAiClient) GenerateUpgradeRationale(pkg, fromVersion, toVersion string) (string, error) {
	prompt := fmt.Sprintf(`Generate a short, technical but clear explanation why upgrading %s from version %s to %s is beneficial.
Focus on security fixes, performance improvements, new features, and bug fixes.
Keep it concise, professional, and factual. Limit to 2-3 sentences.`, pkg, fromVersion, toVersion)

	resp, err := c.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   150,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate rationale: %w", err)
	}

	rationale := strings.TrimSpace(resp.Choices[0].Message.Content)
	return rationale, nil
}

// GeneratePRDescription generates a PR description for multiple package upgrades
func (c *OpenAiClient) GeneratePRDescription(upgrades []Upgrade) (string, error) {
	var upgradesText strings.Builder
	for i, upgrade := range upgrades {
		upgradesText.WriteString(fmt.Sprintf("%d. %s: %s → %s\n", i+1, upgrade.Package, upgrade.FromVersion, upgrade.ToVersion))
		if upgrade.Rationale != "" {
			upgradesText.WriteString(fmt.Sprintf("   Rationale: %s\n", upgrade.Rationale))
		}
	}

	prompt := fmt.Sprintf(`Generate a professional Pull Request description for upgrading these dependencies:

%s

Include:
1. A concise PR title
2. A summary explaining the purpose of these upgrades
3. Any potential breaking changes to watch for
4. Testing recommendations for the changes

Keep it under 300 words total.`, upgradesText.String())

	resp, err := c.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   500,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate PR description: %w", err)
	}

	description := strings.TrimSpace(resp.Choices[0].Message.Content)
	return description, nil
}

// IsCompatibleVersion checks if a package version is compatible with a constraint
func (c *OpenAiClient) IsCompatibleVersion(version, constraint string) (bool, error) {
	return semver.IsCompatible(version, constraint)
}
