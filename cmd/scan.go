package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/dai/pkg/parser"
	"github.com/your-org/dai/pkg/security"
)

func init() {
	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the current project for outdated dependencies",
	Long:  `Scan command parses your package.json (and lockfile) to list direct and transitive dependencies with current versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		scanProject()
	},
}

func scanProject() {
	fmt.Println("Scanning project dependencies...")

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

	// Print dependencies
	if len(pkg.Dependencies) > 0 {
		fmt.Println("Dependencies:")
		for name, version := range pkg.Dependencies {
			fmt.Printf("  - %s: %s\n", name, version)
			checkVulnerabilities(name, version)
		}
		fmt.Println()
	}

	// Print dev dependencies
	if len(pkg.DevDependencies) > 0 {
		fmt.Println("Dev Dependencies:")
		for name, version := range pkg.DevDependencies {
			fmt.Printf("  - %s: %s\n", name, version)
			checkVulnerabilities(name, version)
		}
		fmt.Println()
	}
}

func checkVulnerabilities(name, version string) {
	scanner := security.NewVulnerabilityScanner()
	vulns, err := scanner.ScanPackage(name, version)
	if err != nil || len(vulns) == 0 {
		return
	}

	fmt.Printf("    ⚠️  %d vulnerabilities found!\n", len(vulns))
	// TODO: Display vulnerability details
}
