package cmd

import (
	"fmt"
	"os"

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
		scanProject()
	},
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
