package cmd

import (
	"fmt"
	"os"

	"github.com/HibiZA/dai/pkg/parser"
	"github.com/HibiZA/dai/pkg/security"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	scanDev      bool
)

func init() {
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text or table")
	scanCmd.Flags().BoolVarP(&scanDev, "dev", "d", true, "Scan dev dependencies")
	rootCmd.AddCommand(scanCmd)
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
	fmt.Println("Scanning project dependencies for vulnerabilities...")

	// Find package.json
	dir, err := parser.FindPackageJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse package.json
	pkg, err := parser.ParsePackageJSON(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project: %s@%s\n\n", pkg.Name, pkg.Version)

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
		fmt.Println("No dependencies found to scan.")
		return
	}

	// Create a vulnerability reporter and scan all packages
	scanner := security.NewVulnerabilityScanner()
	reporter := security.NewVulnerabilityReporter(scanner)

	// Generate reports for all packages
	reports, err := reporter.ReportMultiple(packagesToScan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: some packages couldn't be scanned: %v\n\n", err)
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

	fmt.Printf("    ⚠️  %d vulnerabilities found!\n", len(vulns))

	// Generate a report for this package
	reporter := security.NewVulnerabilityReporter(scanner)
	report, err := reporter.GenerateReport(name, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		return
	}

	// Print vulnerability details
	for i, vuln := range report.Vulnerabilities {
		fmt.Printf("      [%d] %s - %s (%s)\n", i+1, vuln.ID, vuln.Description, vuln.Severity)
	}
}
