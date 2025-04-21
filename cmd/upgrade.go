package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/dai/pkg/ai"
	"github.com/your-org/dai/pkg/npm"
	"github.com/your-org/dai/pkg/parser"
	"github.com/your-org/dai/pkg/semver"
)

var (
	allFlag         bool
	createPRFlag    bool
	openaiAPIKey    string
	githubToken     string
	debugFlag       bool
	applyFlag       bool
	dryRunFlag      bool
	registryURLFlag string
	simulateFlag    bool
)

func init() {
	upgradeCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Upgrade all dependencies")
	upgradeCmd.Flags().BoolVarP(&createPRFlag, "pr", "p", false, "Create a pull request with the changes")
	upgradeCmd.Flags().StringVarP(&openaiAPIKey, "openai-key", "k", "", "OpenAI API key (or set DAI_OPENAI_API_KEY env var)")
	upgradeCmd.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub token (or set DAI_GITHUB_TOKEN env var)")
	upgradeCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug output")
	upgradeCmd.Flags().BoolVar(&applyFlag, "apply", false, "Apply upgrades to package.json")
	upgradeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be upgraded without making changes")
	upgradeCmd.Flags().StringVar(&registryURLFlag, "registry", "", "NPM registry URL (defaults to https://registry.npmjs.org)")
	upgradeCmd.Flags().BoolVar(&simulateFlag, "simulate", false, "Simulate upgrades (don't actually check npm registry)")

	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [packages]",
	Short: "Upgrade specific packages and preview diff",
	Long:  `Upgrade command applies version bumps to specified packages and generates a patch with AI-drafted rationales.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Get API keys from flags or environment variables
		if openaiAPIKey == "" {
			openaiAPIKey = os.Getenv("DAI_OPENAI_API_KEY")
		}

		if githubToken == "" {
			githubToken = os.Getenv("DAI_GITHUB_TOKEN")
		}

		var packages []string
		if allFlag {
			packages = []string{"--all"}
		} else if len(args) > 0 {
			packages = strings.Split(args[0], ",")
		}

		upgradePackages(packages)
	},
}

func upgradePackages(packages []string) {
	if len(packages) == 0 && !allFlag {
		fmt.Println("No packages specified, use comma-separated list or --all flag")
		return
	}

	// Find and parse package.json
	dir, err := parser.FindPackageJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	pkg, err := parser.ParsePackageJSON(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine which packages to upgrade
	var packagesToUpgrade map[string]string
	if allFlag || (len(packages) == 1 && packages[0] == "--all") {
		// Upgrade all dependencies
		packagesToUpgrade = make(map[string]string)
		for name, version := range pkg.Dependencies {
			packagesToUpgrade[name] = version
		}
		for name, version := range pkg.DevDependencies {
			packagesToUpgrade[name] = version
		}
	} else {
		// Upgrade only specified packages
		packagesToUpgrade = make(map[string]string)
		for _, name := range packages {
			if version, ok := pkg.Dependencies[name]; ok {
				packagesToUpgrade[name] = version
			} else if version, ok := pkg.DevDependencies[name]; ok {
				packagesToUpgrade[name] = version
			} else {
				fmt.Printf("Warning: Package %s not found in dependencies\n", name)
			}
		}
	}

	if len(packagesToUpgrade) == 0 {
		fmt.Println("No valid packages to upgrade")
		return
	}

	fmt.Printf("Checking upgrades for %d packages...\n", len(packagesToUpgrade))

	// Add debug output for semver parsing
	if debugFlag {
		fmt.Println("\nDebug: Semver parsing")
		for name, version := range packagesToUpgrade {
			fmt.Printf("Package %s, Version: %s\n", name, version)

			// Try to match with constraint regex
			matches := semver.ConstraintRegexForDebug().FindStringSubmatch(version)
			if matches == nil {
				fmt.Printf("  Failed to match constraint regex\n")
			} else {
				fmt.Printf("  Constraint regex matches: %v\n", matches)
			}

			// Try to extract version
			if matches != nil && len(matches) > 2 {
				versionStr := matches[2]
				fmt.Printf("  Extracted version: %s\n", versionStr)

				// Try to match with semver regex
				semverMatches := semver.SemverRegexForDebug().FindStringSubmatch(versionStr)
				if semverMatches == nil {
					fmt.Printf("  Failed to match semver regex\n")
				} else {
					fmt.Printf("  Semver regex matches: %v\n", semverMatches)
				}
			}
		}
		fmt.Println()
	}

	// Create npm registry client unless we're in simulate mode
	var registryClient *npm.RegistryClient
	if !simulateFlag {
		registryClient = npm.NewRegistryClient(registryURLFlag)
	}

	// Find upgrades for each package
	upgrades := make(map[string]ai.VersionUpgrade)

	for name, versionConstraint := range packagesToUpgrade {
		var newVersionStr string

		if simulateFlag {
			// Parse the constraint and extract the actual version for simulation
			constraintType, currentVersion, err := semver.ParseConstraint(versionConstraint)
			if err != nil {
				fmt.Printf("Warning: Could not parse version for %s: %v\n", name, err)
				continue
			}

			// Simulate a bumped version
			newVersion := &semver.Version{
				Major: currentVersion.Major,
				Minor: currentVersion.Minor + 1,
				Patch: 0,
			}

			// Format with the original constraint
			newVersionStr = constraintType + newVersion.String()
		} else {
			// Use the npm registry to find the best upgrade
			var err error
			newVersionStr, err = registryClient.FindBestUpgrade(name, versionConstraint)
			if err != nil {
				fmt.Printf("Warning: No upgrade found for %s: %v\n", name, err)
				continue
			}
		}

		// Skip if the version didn't change
		if newVersionStr == versionConstraint {
			fmt.Printf("No updates available for %s (already at latest: %s)\n", name, versionConstraint)
			continue
		}

		fmt.Printf("Upgrade found for %s: %s → %s\n", name, versionConstraint, newVersionStr)

		// Extract versions without constraints for AI
		_, currentVersion, _ := semver.ParseConstraint(versionConstraint)
		_, newVersion, _ := semver.ParseConstraint(newVersionStr)

		// Get AI-generated rationale
		aiClient := ai.NewOpenAiClient(openaiAPIKey)
		rationale, err := aiClient.GenerateUpgradeRationale(name, currentVersion.String(), newVersion.String(), []string{})
		if err != nil {
			fmt.Printf("Warning: Failed to generate rationale: %v\n", err)
		}

		upgrades[name] = ai.VersionUpgrade{
			PackageName: name,
			OldVersion:  versionConstraint,
			NewVersion:  newVersionStr,
			Rationale:   rationale,
			Breaking:    false, // We would need to analyze the changes to determine this
		}
	}

	if len(upgrades) == 0 {
		fmt.Println("\nNo upgrades found for the specified packages.")
		return
	}

	// Apply upgrades if requested
	if applyFlag && !dryRunFlag && len(upgrades) > 0 {
		fmt.Println("\nApplying upgrades to package.json...")

		// Apply each upgrade
		for name, upgrade := range upgrades {
			updated := pkg.UpdateDependency(name, upgrade.NewVersion)
			if !updated {
				fmt.Printf("Warning: Failed to update %s\n", name)
			}
		}

		// Write the updated file
		if err := pkg.WriteToFile(dir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to write package.json: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Upgrades applied successfully!")
	} else if dryRunFlag {
		fmt.Println("\nDry run - no changes were made to package.json")
	} else if !applyFlag {
		fmt.Println("\nUse --apply to write changes to package.json")
	}

	// Generate PR description
	if createPRFlag && len(upgrades) > 0 {
		aiClient := ai.NewOpenAiClient(openaiAPIKey)
		description, err := aiClient.GeneratePRDescription(upgrades)
		if err != nil {
			fmt.Printf("Warning: Failed to generate PR description: %v\n", err)
		} else {
			fmt.Println("\nPR Description Preview:")
			fmt.Println("------------------------")
			fmt.Println(description)
			fmt.Println("------------------------")

			// Create PR if GitHub token is provided
			if githubToken != "" {
				// TODO: Implement GitHub PR creation
				fmt.Println("\nCreating pull request...")
				// For now, just print a message
				fmt.Println("PR creation not implemented yet")
			} else {
				fmt.Println("\nSkipping PR creation (no GitHub token provided)")
			}
		}
	}
}
