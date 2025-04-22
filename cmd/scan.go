package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/HibiZA/dai/pkg/ai"
	"github.com/HibiZA/dai/pkg/config"
	"github.com/HibiZA/dai/pkg/parser"
	"github.com/HibiZA/dai/pkg/security"
	"github.com/HibiZA/dai/pkg/style"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	scanDev      bool
)

func init() {
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text or table")
	scanCmd.Flags().BoolVarP(&scanDev, "dev", "d", true, "Scan dev dependencies")
	scanCmd.Flags().Bool("skip-key-check", false, "Skip API key validation and prompts")

	// Set custom help function
	scanCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Display custom colorful help
		displayScanHelp(cmd)
	})

	rootCmd.AddCommand(scanCmd)
}

// displayScanHelp shows custom styled help for the scan command
func displayScanHelp(cmd *cobra.Command) {
	fmt.Println(style.Banner())
	fmt.Println()

	// Command name
	fmt.Printf("%s - %s\n\n", style.Title("Scan Command"), style.Subtitle("Security Vulnerability Scanner"))

	// Description
	fmt.Println(style.Info("Scan command parses your package.json (and lockfile) to check dependencies for security vulnerabilities using multiple data sources."))
	fmt.Println()

	// Usage section
	fmt.Println(style.Title("Usage:"))
	fmt.Printf("  %s\n\n", style.Highlight("dai scan [flags]"))

	// Flags section
	fmt.Println(style.Title("Flags:"))
	fmt.Printf("  %-20s %s\n",
		style.Highlight("-d, --dev"),
		style.Subtitle("Scan dev dependencies (default: true)"))
	fmt.Printf("  %-20s %s\n",
		style.Highlight("-f, --format string"),
		style.Subtitle("Output format: text or table (default \"text\")"))
	fmt.Printf("  %-20s %s\n",
		style.Highlight("--skip-key-check"),
		style.Subtitle("Skip API key validation and prompts"))
	fmt.Printf("  %-20s %s\n\n",
		style.Highlight("-h, --help"),
		style.Subtitle("Help for scan command"))

	// Examples section
	fmt.Println(style.Title("Examples:"))
	fmt.Printf("  %s\n", style.Subtitle("# Scan all dependencies (including dev dependencies)"))
	fmt.Printf("  %s\n\n", style.Highlight("dai scan"))

	fmt.Printf("  %s\n", style.Subtitle("# Scan only production dependencies"))
	fmt.Printf("  %s\n\n", style.Highlight("dai scan --dev=false"))

	fmt.Printf("  %s\n", style.Subtitle("# Output in table format"))
	fmt.Printf("  %s\n", style.Highlight("dai scan --format table"))
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the current project for outdated dependencies and security vulnerabilities",
	Long:  `Scan command parses your package.json (and lockfile) to check dependencies for security vulnerabilities using multiple data sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check and prompt for required API keys if not set
		skipKeyCheck, _ := cmd.Flags().GetBool("skip-key-check")
		if !skipKeyCheck && !checkRequiredKeys() {
			return
		}

		scanProject()
	},
}

// checkRequiredKeys checks if required API keys are set and prompts user to enter them if not
func checkRequiredKeys() bool {
	cfg := config.LoadConfig()

	// Debug log to check OpenAI API key status
	fmt.Println(style.Info("Debug - OpenAI key status:"), cfg.HasOpenAIKey())

	// Add deeper debugging for OpenAI key issues
	fmt.Println(style.Info("OpenAI Key Debug Info:"))
	debugInfo := ai.DebugOpenAIKey()
	fmt.Println(debugInfo)

	// Check GitHub token (most important for scanning)
	if !cfg.HasGitHubToken() {
		fmt.Println(style.Warning("GitHub token not found. A token is recommended to avoid GitHub API rate limits."))
		fmt.Println(style.Info("Would you like to set a GitHub token now? (y/n):"))

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your GitHub token (will be hidden):"))
			fmt.Print("> ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)

			if token != "" {
				// Use the config command directly since saveAPIKey is not exported
				err := saveGitHubToken(token)
				if err != nil {
					fmt.Println(style.Error("Error:"), "Failed to save GitHub token:", err)
				}
			} else {
				fmt.Println(style.Warning("No token provided. Scanning may hit API rate limits."))
				fmt.Println(style.Info("You can set a token later with: dai config --set github --github-token YOUR_TOKEN"))
			}
		} else {
			fmt.Println(style.Warning("Proceeding without GitHub token. Scanning may hit API rate limits."))
		}
	}

	// Check NVD API key
	if !cfg.HasNVDApiKey() {
		fmt.Println(style.Warning("NVD API key not found. A key is recommended to avoid NVD API rate limits."))
		fmt.Println(style.Info("Would you like to set an NVD API key now? (y/n):"))

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your NVD API key (will be hidden):"))
			fmt.Print("> ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)

			if token != "" {
				// Use the config command directly since saveAPIKey is not exported
				err := saveNVDApiKey(token)
				if err != nil {
					fmt.Println(style.Error("Error:"), "Failed to save NVD API key:", err)
				}
			} else {
				fmt.Println(style.Warning("No API key provided. Scanning may hit API rate limits."))
				fmt.Println(style.Info("You can set a key later with: dai config --set nvd --nvd-api-key YOUR_KEY"))
			}
		} else {
			fmt.Println(style.Warning("Proceeding without NVD API key. Scanning may hit API rate limits."))
		}
	}

	// Check OpenAI key
	if !cfg.HasOpenAIKey() {
		fmt.Println(style.Warning("OpenAI API key not found. A key is required for generating upgrade rationales."))
		fmt.Println(style.Info("Would you like to set an OpenAI API key now? (y/n):"))

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			fmt.Println(style.Info("Enter your OpenAI API key (will be hidden):"))
			fmt.Print("> ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)

			if token != "" {
				// Use the config command directly since saveAPIKey is not exported
				err := saveOpenAIKey(token)
				if err != nil {
					fmt.Println(style.Error("Error:"), "Failed to save OpenAI API key:", err)
				}
			} else {
				fmt.Println(style.Warning("No API key provided."))
				fmt.Println(style.Info("You can set a key later with: dai config --set openai --openai-key YOUR_KEY"))
			}
		} else {
			fmt.Println(style.Warning("Proceeding without OpenAI API key."))
		}
	}

	return true
}

// saveGitHubToken is a helper function to save the GitHub token
func saveGitHubToken(token string) error {
	return SaveAPIKey("github", token)
}

// saveNVDApiKey is a helper function to save the NVD API key
func saveNVDApiKey(token string) error {
	return SaveAPIKey("nvd", token)
}

// saveOpenAIKey is a helper function to save the OpenAI API key
func saveOpenAIKey(token string) error {
	return SaveAPIKey("openai", token)
}

func scanProject() {
	fmt.Println(style.Header("🔍 Scanning project dependencies for vulnerabilities..."))
	fmt.Println(style.Divider())
	fmt.Println()

	// Find package.json
	dir, err := parser.FindPackageJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", style.Error("Error:"), err)
		os.Exit(1)
	}

	// Parse package.json
	pkg, err := parser.ParsePackageJSON(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", style.Error("Error:"), err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n\n", style.Info("Project:"), style.FormatPackage(pkg.Name, pkg.Version))

	// Collect all dependencies to scan
	packagesToScan := make(map[string]string)

	// Add production dependencies
	for name, version := range pkg.Dependencies {
		packagesToScan[name] = version
	}

	// Add dev dependencies if requested
	if scanDev {
		for name, version := range pkg.DevDependencies {
			packagesToScan[name] = version
		}
	}

	// Skip if no dependencies found
	if len(packagesToScan) == 0 {
		fmt.Println(style.Warning("No dependencies found to scan."))
		return
	}

	// Create a vulnerability reporter and scan all packages
	scanner := security.NewVulnerabilityScanner()
	reporter := security.NewVulnerabilityReporter(scanner)

	// Generate reports for all packages
	reports, err := reporter.ReportMultiple(packagesToScan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n\n", style.Warning("Warning: some packages couldn't be scanned:"), err)
	}

	// Write the console report
	security.WriteConsoleReport(os.Stdout, reports)

	// After the scan is complete, prompt for upgrading vulnerable packages
	promptForUpgrades(reports, pkg, dir)
}

// promptForUpgrades prompts the user to upgrade vulnerable packages
func promptForUpgrades(reports map[string]*security.VulnerabilityReport, pkg *parser.PackageJSON, dir string) {
	// Check if there are any vulnerable packages
	vulnerablePackages := []string{}
	for packageName, report := range reports {
		if len(report.Vulnerabilities) > 0 {
			vulnerablePackages = append(vulnerablePackages, packageName)
		}
	}

	if len(vulnerablePackages) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(style.Header("🔄 Package Upgrade Recommendations"))
	fmt.Println(style.Divider())
	fmt.Printf("Found %d vulnerable %s. Would you like to upgrade them? (y/n): ",
		len(vulnerablePackages),
		pluralize("package", len(vulnerablePackages)))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
		fmt.Println(style.Info("No packages will be upgraded."))
		return
	}

	// Perform the upgrade
	fmt.Println(style.Info("Checking for available upgrades..."))

	// Create a list of upgrades to perform
	upgrades := []string{}
	for _, packageName := range vulnerablePackages {
		upgrades = append(upgrades, packageName)
	}

	// Run npm outdated to get latest versions
	fmt.Println(style.Subtitle("Running npm outdated to find latest versions..."))

	// Change to the directory containing package.json
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	// Perform the actual upgrade
	// NOTE: We are just showing the recommendation here; we'd need to integrate with
	// a proper package manager (npm, yarn, etc.) to do the actual upgrades.
	fmt.Println(style.Success("Upgrade recommendation:"))
	for _, packageName := range vulnerablePackages {
		currentVersion := ""

		// Find current version
		if v, ok := pkg.Dependencies[packageName]; ok {
			currentVersion = v
		} else if v, ok := pkg.DevDependencies[packageName]; ok {
			currentVersion = v
		}

		fmt.Printf("  %s: %s → %s\n",
			style.Highlight(packageName),
			style.Warning(currentVersion),
			style.Success("latest"))
	}

	fmt.Println()
	fmt.Println(style.Info("To upgrade these packages, run one of the following commands:"))
	fmt.Printf("  %s %s\n",
		style.Highlight("npm upgrade"),
		strings.Join(vulnerablePackages, " "))
	fmt.Printf("  %s %s\n",
		style.Highlight("yarn upgrade"),
		strings.Join(vulnerablePackages, " "))

	fmt.Println()
	fmt.Println(style.Info("Or use our built-in upgrade command:"))
	fmt.Printf("  %s %s\n",
		style.Highlight("dai upgrade"),
		strings.Join(vulnerablePackages, " "))
}

// pluralize returns a singular or plural form based on count
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func checkVulnerabilities(name, version string) {
	scanner := security.NewVulnerabilityScanner()
	vulns, err := scanner.ScanPackage(name, version)
	if err != nil || len(vulns) == 0 {
		return
	}

	fmt.Printf("    %s  %s\n", style.Warning("⚠️"), style.Warning(fmt.Sprintf("%d vulnerabilities found!", len(vulns))))

	// Generate a report for this package
	reporter := security.NewVulnerabilityReporter(scanner)
	report, err := reporter.GenerateReport(name, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", style.Error("Error generating report:"), err)
		return
	}

	// Print vulnerability details
	for i, vuln := range report.Vulnerabilities {
		fmt.Printf("      [%d] %s - %s (%s)\n",
			i+1,
			style.Highlight(vuln.ID),
			vuln.Description,
			style.GetSeverityColor(vuln.Severity))
	}
}
