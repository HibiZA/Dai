package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HibiZA/dai/pkg/ai"
	"github.com/HibiZA/dai/pkg/config"
	"github.com/HibiZA/dai/pkg/github"
	"github.com/HibiZA/dai/pkg/npm"
	"github.com/HibiZA/dai/pkg/parser"
	"github.com/HibiZA/dai/pkg/semver"
	"github.com/HibiZA/dai/pkg/style"
	"github.com/spf13/cobra"
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
	testAIFlag      bool
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
	upgradeCmd.Flags().BoolVar(&testAIFlag, "test-ai", false, "Test AI-generated content quality for specific packages")
	upgradeCmd.Flags().Bool("skip-key-check", false, "Skip API key validation and prompts")

	// Set custom help function
	upgradeCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Display custom colorful help
		displayUpgradeHelp(cmd)
	})

	rootCmd.AddCommand(upgradeCmd)
}

// displayUpgradeHelp shows custom styled help for the upgrade command
func displayUpgradeHelp(cmd *cobra.Command) {
	fmt.Println(style.Banner())
	fmt.Println()

	// Command name
	fmt.Printf("%s - %s\n\n", style.Title("Upgrade Command"), style.Subtitle("Smart Dependency Upgrader"))

	// Description
	fmt.Println(style.Info("Upgrade command applies version bumps to specified packages and generates a patch with AI-drafted rationales."))
	fmt.Println()

	// Usage section
	fmt.Println(style.Title("Usage:"))
	fmt.Printf("  %s\n\n", style.Highlight("dai upgrade [packages] [flags]"))

	// Flags section
	fmt.Println(style.Title("Flags:"))

	// Format all flags in a consistent and colorful way
	flagsInfo := []struct {
		flag string
		desc string
	}{
		{"-a, --all", "Upgrade all dependencies"},
		{"--apply", "Apply upgrades to package.json"},
		{"-d, --debug", "Enable debug output"},
		{"--dry-run", "Show what would be upgraded without making changes"},
		{"-k, --openai-key string", "OpenAI API key (or set DAI_OPENAI_API_KEY env var)"},
		{"-p, --pr", "Create a pull request with the changes"},
		{"--registry string", "NPM registry URL (defaults to https://registry.npmjs.org)"},
		{"--simulate", "Simulate upgrades (don't actually check npm registry)"},
		{"--skip-key-check", "Skip API key validation and prompts"},
		{"--test-ai", "Test AI-generated content quality for specific packages"},
		{"-t, --github-token string", "GitHub token (or set DAI_GITHUB_TOKEN env var)"},
		{"-h, --help", "Help for upgrade command"},
	}

	for _, f := range flagsInfo {
		fmt.Printf("  %-32s %s\n",
			style.Highlight(f.flag),
			style.Subtitle(f.desc))
	}
	fmt.Println()

	// Examples section
	fmt.Println(style.Title("Examples:"))
	fmt.Printf("  %s\n", style.Subtitle("# Upgrade a specific package"))
	fmt.Printf("  %s\n\n", style.Highlight("dai upgrade react"))

	fmt.Printf("  %s\n", style.Subtitle("# Upgrade multiple packages"))
	fmt.Printf("  %s\n\n", style.Highlight("dai upgrade react,react-dom,redux"))

	fmt.Printf("  %s\n", style.Subtitle("# Upgrade all dependencies"))
	fmt.Printf("  %s\n\n", style.Highlight("dai upgrade --all"))

	fmt.Printf("  %s\n", style.Subtitle("# Apply upgrades and create a PR"))
	fmt.Printf("  %s\n", style.Highlight("dai upgrade --all --apply --pr"))
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

		// Check and prompt for required API keys if not set
		skipKeyCheck, _ := cmd.Flags().GetBool("skip-key-check")
		if !skipKeyCheck && !checkRequiredUpgradeKeys() {
			return
		}

		// If test-ai flag is set, run AI testing instead of normal upgrade
		if testAIFlag {
			testAIContentQuality(args)
			return
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

// checkRequiredUpgradeKeys checks if required API keys are set and prompts user to enter them if not
func checkRequiredUpgradeKeys() bool {
	cfg := config.LoadConfig()
	reader := bufio.NewReader(os.Stdin)

	// Check GitHub token
	if !cfg.HasGitHubToken() && createPRFlag {
		fmt.Println(style.Warning("GitHub token is required for creating PRs."))
		fmt.Println(style.Info("Would you like to set a GitHub token now? (y/n):"))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your GitHub token (will be hidden):"))
			fmt.Print("> ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)

			if token != "" {
				if err := SaveAPIKey("github", token); err != nil {
					fmt.Printf("%s %v\n", style.Error("Error:"), err)
					return false
				}
				githubToken = token // Set for current session
			} else {
				fmt.Println(style.Error("No token provided. Cannot create PR."))
				return false
			}
		} else {
			fmt.Println(style.Error("GitHub token is required for PR creation. Aborting."))
			return false
		}
	} else if !cfg.HasGitHubToken() {
		fmt.Println(style.Warning("GitHub token not found. This may affect API rate limits."))
		fmt.Println(style.Info("Would you like to set a GitHub token now? (y/n):"))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your GitHub token:"))
			fmt.Print("> ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)

			if token != "" {
				if err := SaveAPIKey("github", token); err != nil {
					fmt.Printf("%s %v\n", style.Error("Error:"), err)
					return false
				}
				githubToken = token // Set for current session
			}
		}
	}

	// Check OpenAI API key if not using simulation
	if !cfg.HasOpenAIKey() && !simulateFlag {
		fmt.Println(style.Warning("OpenAI API key not found. This is required for AI-generated upgrade rationales."))
		fmt.Println(style.Info("Would you like to set an OpenAI API key now? (y/n):"))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your OpenAI API key:"))
			fmt.Print("> ")
			key, _ := reader.ReadString('\n')
			key = strings.TrimSpace(key)

			if key != "" {
				if err := SaveAPIKey("openai", key); err != nil {
					fmt.Printf("%s %v\n", style.Error("Error:"), err)
					return false
				}
				openaiAPIKey = key // Set for current session
			} else {
				fmt.Println(style.Warning("No key provided. Will proceed without AI-generated rationales."))
			}
		} else {
			fmt.Println(style.Warning("Proceeding without AI-generated rationales."))
		}
	}

	return true
}

func upgradePackages(packages []string) {
	if len(packages) == 0 && !allFlag {
		fmt.Println(style.Warning("No packages specified, use comma-separated list or --all flag"))
		return
	}

	// Find and parse package.json
	dir, err := parser.FindPackageJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", style.Error("Error:"), err)
		os.Exit(1)
	}

	pkg, err := parser.ParsePackageJSON(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", style.Error("Error:"), err)
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
		fmt.Println(style.Warning("No valid packages to upgrade"))
		return
	}

	fmt.Printf("%s %s\n",
		style.Header("🔄 Checking upgrades for"),
		style.Highlight(fmt.Sprintf("%d packages...", len(packagesToUpgrade))))
	fmt.Println(style.Divider())
	fmt.Println()

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
		cfg := &config.Config{OpenAIApiKey: openaiAPIKey}
		aiClient, err := ai.NewOpenAiClient(cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to create OpenAI client: %v\n", err)
			continue
		}
		rationale, err := aiClient.GenerateUpgradeRationale(name, currentVersion.String(), newVersion.String())
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

		// Create backup of package.json first
		backupPath, err := parser.CreateBackup(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create backup of package.json: %v\n", err)
			os.Exit(1)
		}

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

		// Generate diff between the original and modified file
		diff, err := parser.GenerateDiff(dir, backupPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to generate diff: %v\n", err)
		} else {
			fmt.Println("\nChanges to package.json:")
			fmt.Println("------------------------")
			fmt.Println(diff)
			fmt.Println("------------------------")
		}

		fmt.Println("Upgrades applied successfully!")
	} else if dryRunFlag {
		fmt.Println("\nDry run - no changes were made to package.json")
	} else if !applyFlag {
		fmt.Println("\nUse --apply to write changes to package.json")
	}

	// Generate PR description
	if createPRFlag && len(upgrades) > 0 {
		cfg := &config.Config{OpenAIApiKey: openaiAPIKey}
		aiClient, err := ai.NewOpenAiClient(cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to create OpenAI client: %v\n", err)
			return
		}

		// Convert map[string]VersionUpgrade to []Upgrade
		upgradesList := make([]ai.Upgrade, 0, len(upgrades))
		for _, u := range upgrades {
			upgradesList = append(upgradesList, ai.Upgrade{
				Package:     u.PackageName,
				FromVersion: u.OldVersion,
				ToVersion:   u.NewVersion,
				Rationale:   u.Rationale,
			})
		}

		description, err := aiClient.GeneratePRDescription(upgradesList)
		if err != nil {
			fmt.Printf("Warning: Failed to generate PR description: %v\n", err)
		} else {
			// Capture diff for PR description
			var patchInfo string
			if applyFlag && !dryRunFlag {
				// Get the diff from the backup
				diff, err := parser.GenerateDiff(dir, filepath.Join(dir, "package.json.bak"))
				if err == nil {
					patchInfo = "\n\n## Changes\n\n```diff\n" + diff + "\n```"
				}
			}

			// Append diff to description if available
			if patchInfo != "" {
				description += patchInfo
			}

			fmt.Println("\nPR Description Preview:")
			fmt.Println("------------------------")
			fmt.Println(description)
			fmt.Println("------------------------")

			// Create PR if GitHub token is provided
			if githubToken != "" && applyFlag && !dryRunFlag {
				fmt.Println("\nCreating pull request...")

				// Try to determine repository owner and name
				owner, repo, err := github.GetRepoDetails()
				if err != nil {
					fmt.Printf("Warning: Could not determine repository details: %v\n", err)
					fmt.Println("Please provide repository details manually")
					fmt.Print("Owner: ")
					fmt.Scanln(&owner)
					fmt.Print("Repository: ")
					fmt.Scanln(&repo)
				}

				if owner == "" || repo == "" {
					fmt.Println("Error: Repository details required for PR creation")
					return
				}

				// Initialize GitHub client
				githubClient, err := github.NewGitHubClient(githubToken, owner, repo)
				if err != nil {
					fmt.Printf("Error: Failed to create GitHub client: %v\n", err)
					return
				}

				// Create a unique branch name based on timestamp
				branchName := fmt.Sprintf("dai-dependency-update-%d", time.Now().Unix())

				// Get the default branch to use as base
				baseBranch, err := githubClient.GetDefaultBranch()
				if err != nil {
					fmt.Printf("Warning: Could not determine default branch: %v\n", err)
					baseBranch = "main" // Fallback to main
				}

				// Read the content of the modified package.json
				packageJSONPath := filepath.Join(dir, "package.json")
				packageJSONContent, err := os.ReadFile(packageJSONPath)
				if err != nil {
					fmt.Printf("Error: Failed to read package.json: %v\n", err)
					return
				}

				// Files to modify in the PR
				files := map[string]string{
					"package.json": string(packageJSONContent),
				}

				// Prepare the PR details
				prTitle := "chore(deps): update dependencies"
				if len(upgrades) == 1 {
					// If only one package was upgraded, make the title more specific
					for name, upgrade := range upgrades {
						prTitle = fmt.Sprintf("chore(deps): update %s from %s to %s",
							name, upgrade.OldVersion, upgrade.NewVersion)
						break
					}
				} else {
					// Multiple packages upgraded
					prTitle = fmt.Sprintf("chore(deps): update %d dependencies", len(upgrades))
				}

				pr := &github.PullRequest{
					Title:       prTitle,
					Description: description,
					Branch:      branchName,
					BaseBranch:  baseBranch,
					Files:       files,
				}

				// Create the pull request
				prURL, err := githubClient.CreatePullRequest(pr)
				if err != nil {
					fmt.Printf("Error: Failed to create pull request: %v\n", err)
					return
				}

				fmt.Printf("Pull request created successfully: %s\n", prURL)
			} else if githubToken == "" {
				fmt.Println("\nSkipping PR creation (no GitHub token provided)")
			} else if !applyFlag || dryRunFlag {
				fmt.Println("\nSkipping PR creation (changes not applied to package.json)")
			}
		}
	}

	// Show the upgrades found
	if len(upgrades) == 0 {
		fmt.Println(style.Success("✅ All packages are already up to date!"))
		return
	}

	// Display upgrade summary
	fmt.Printf("%s\n", style.Title("📦 Dependency Upgrade Summary"))
	fmt.Println(style.Divider())
	fmt.Printf("%s %d\n", style.Info("Total packages checked:"), len(upgrades))
	fmt.Printf("%s %s\n\n", style.Success("Packages to upgrade:"), style.Success(fmt.Sprintf("%d", len(upgrades))))

	// Print table header
	fmt.Printf("%-20s %-15s %-15s %s\n",
		style.Header("Package"),
		style.Header("Current"),
		style.Header("New Version"),
		style.Header("Update Type"))
	fmt.Println(strings.Repeat("-", 80))

	// ... rest of the function ...
}

// testAIContentQuality tests the quality of AI-generated content for packages
func testAIContentQuality(args []string) {
	if openaiAPIKey == "" {
		fmt.Println("Error: OpenAI API key is required for AI testing. Set it with --openai-key flag or DAI_OPENAI_API_KEY env var.")
		return
	}

	// Test packages - use defaults if none specified
	testPackages := []struct {
		name       string
		oldVersion string
		newVersion string
	}{
		{"react", "16.14.0", "18.2.0"},
		{"express", "4.17.1", "4.18.2"},
		{"lodash", "4.17.20", "4.17.21"},
	}

	// Allow custom packages to be specified
	if len(args) > 0 && args[0] != "--all" {
		customPackages := strings.Split(args[0], ",")
		if len(customPackages) > 0 {
			// If user provides custom packages, use those instead
			testPackages = nil
			for _, pkg := range customPackages {
				// Use placeholder versions for custom packages
				testPackages = append(testPackages, struct {
					name       string
					oldVersion string
					newVersion string
				}{
					name:       pkg,
					oldVersion: "1.0.0",
					newVersion: "2.0.0",
				})
			}
		}
	}

	// Create OpenAI client
	cfg := &config.Config{OpenAIApiKey: openaiAPIKey}
	aiClient, err := ai.NewOpenAiClient(cfg)
	if err != nil {
		fmt.Printf("Error: Failed to create OpenAI client: %v\n", err)
		return
	}

	fmt.Println("Testing AI-generated content quality...")
	fmt.Println("======================================")

	// Test rationale generation
	fmt.Println("\n1. Testing Upgrade Rationale Generation:")
	fmt.Println("----------------------------------------")
	for _, pkg := range testPackages {
		fmt.Printf("\nPackage: %s (v%s → v%s)\n", pkg.name, pkg.oldVersion, pkg.newVersion)

		rationale, err := aiClient.GenerateUpgradeRationale(pkg.name, pkg.oldVersion, pkg.newVersion)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Printf("  Rationale: %s\n", rationale)
	}

	// Test PR description generation
	fmt.Println("\n2. Testing PR Description Generation:")
	fmt.Println("-----------------------------------")
	var upgrades []ai.Upgrade
	for _, pkg := range testPackages {
		// Generate rationale for each package
		rationale, _ := aiClient.GenerateUpgradeRationale(pkg.name, pkg.oldVersion, pkg.newVersion)

		upgrades = append(upgrades, ai.Upgrade{
			Package:     pkg.name,
			FromVersion: pkg.oldVersion,
			ToVersion:   pkg.newVersion,
			Rationale:   rationale,
		})
	}

	description, err := aiClient.GeneratePRDescription(upgrades)
	if err != nil {
		fmt.Printf("Error generating PR description: %v\n", err)
	} else {
		fmt.Println(description)
	}

	fmt.Println("\nAI content quality testing complete!")
}
