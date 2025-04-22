package security

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/HibiZA/dai/pkg/style"
)

// SeverityWeight maps severity levels to numeric weights for sorting
var SeverityWeight = map[string]int{
	"CRITICAL": 4,
	"HIGH":     3,
	"MEDIUM":   2,
	"LOW":      1,
	"UNKNOWN":  0,
}

// VulnerabilityReport represents a security report for a package
type VulnerabilityReport struct {
	Package         string
	Version         string
	Vulnerabilities []Vulnerability
	Timestamp       time.Time
}

// NewVulnerabilityReport creates a new vulnerability report
func NewVulnerabilityReport(packageName, version string, vulns []Vulnerability) *VulnerabilityReport {
	return &VulnerabilityReport{
		Package:         packageName,
		Version:         version,
		Vulnerabilities: vulns,
		Timestamp:       time.Now(),
	}
}

// HasVulnerabilities returns true if the report contains any vulnerabilities
func (r *VulnerabilityReport) HasVulnerabilities() bool {
	return len(r.Vulnerabilities) > 0
}

// CountBySeverity returns the count of vulnerabilities by severity level
func (r *VulnerabilityReport) CountBySeverity() map[string]int {
	counts := make(map[string]int)
	for _, vuln := range r.Vulnerabilities {
		severity := strings.ToUpper(vuln.Severity)
		counts[severity]++
	}
	return counts
}

// SortBySeverity sorts the vulnerabilities by severity (most severe first)
func (r *VulnerabilityReport) SortBySeverity() {
	sort.SliceStable(r.Vulnerabilities, func(i, j int) bool {
		// First sort by severity weight (higher first)
		severityI := strings.ToUpper(r.Vulnerabilities[i].Severity)
		severityJ := strings.ToUpper(r.Vulnerabilities[j].Severity)
		weightI := SeverityWeight[severityI]
		weightJ := SeverityWeight[severityJ]

		if weightI != weightJ {
			return weightI > weightJ
		}

		// If same severity, sort by published date (newer first)
		return r.Vulnerabilities[i].Published.After(r.Vulnerabilities[j].Published)
	})
}

// WriteText writes a text-based report to the provided writer
func (r *VulnerabilityReport) WriteText(w io.Writer) error {
	// Sort vulnerabilities by severity
	r.SortBySeverity()

	// Write header
	fmt.Fprintf(w, "Security Vulnerability Report for %s@%s\n", r.Package, r.Version)
	fmt.Fprintf(w, "Generated: %s\n\n", r.Timestamp.Format(time.RFC1123))

	// If no vulnerabilities, print a clean bill of health
	if !r.HasVulnerabilities() {
		fmt.Fprintf(w, "✅ No vulnerabilities found\n")
		return nil
	}

	// Write summary
	fmt.Fprintf(w, "Found %d vulnerabilities:\n", len(r.Vulnerabilities))
	counts := r.CountBySeverity()
	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}
	for _, sev := range severities {
		if count, ok := counts[sev]; ok && count > 0 {
			fmt.Fprintf(w, "  • %s: %d\n", sev, count)
		}
	}
	fmt.Fprintln(w)

	// Write detailed vulnerability information
	fmt.Fprintln(w, "Vulnerability Details:")
	fmt.Fprintln(w, "-----------------------")

	for i, vuln := range r.Vulnerabilities {
		fmt.Fprintf(w, "[%d] %s (%s)\n", i+1, vuln.ID, vuln.Severity)
		fmt.Fprintf(w, "    Description: %s\n", vuln.Description)
		fmt.Fprintf(w, "    Published: %s\n", vuln.Published.Format("2006-01-02"))
		if len(vuln.References) > 0 {
			fmt.Fprintf(w, "    References:\n")
			for _, ref := range vuln.References {
				fmt.Fprintf(w, "      • %s\n", ref)
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}

// WriteTable writes a tabular report to the provided writer
func (r *VulnerabilityReport) WriteTable(w io.Writer) error {
	// Sort vulnerabilities by severity
	r.SortBySeverity()

	// Create a tabwriter
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Write header
	fmt.Fprintf(tw, "ID\tSeverity\tPublished\tDescription\n")
	fmt.Fprintf(tw, "----\t--------\t---------\t-----------\n")

	// Write each vulnerability
	for _, vuln := range r.Vulnerabilities {
		// Truncate description if too long
		desc := vuln.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			vuln.ID,
			vuln.Severity,
			vuln.Published.Format("2006-01-02"),
			desc)
	}

	tw.Flush()
	return nil
}

// VulnerabilityReporter provides methods to report on vulnerabilities
type VulnerabilityReporter struct {
	scanner VulnerabilityScanner
}

// NewVulnerabilityReporter creates a new vulnerability reporter
func NewVulnerabilityReporter(scanner VulnerabilityScanner) *VulnerabilityReporter {
	return &VulnerabilityReporter{
		scanner: scanner,
	}
}

// GenerateReport creates a vulnerability report for a package
func (r *VulnerabilityReporter) GenerateReport(packageName, version string) (*VulnerabilityReport, error) {
	// Scan for vulnerabilities
	vulnerabilities, err := r.scanner.ScanPackage(packageName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to scan package: %w", err)
	}

	// Create and return the report
	return NewVulnerabilityReport(packageName, version, vulnerabilities), nil
}

// ReportMultiple generates reports for multiple packages
func (r *VulnerabilityReporter) ReportMultiple(packages map[string]string) (map[string]*VulnerabilityReport, error) {
	reports := make(map[string]*VulnerabilityReport)
	var errors []error

	for pkg, version := range packages {
		report, err := r.GenerateReport(pkg, version)
		if err != nil {
			errors = append(errors, fmt.Errorf("error scanning %s@%s: %w", pkg, version, err))
			continue
		}
		reports[pkg] = report
	}

	if len(errors) > 0 {
		// Return the reports we did generate, but also return an error
		errMsgs := make([]string, len(errors))
		for i, err := range errors {
			errMsgs[i] = err.Error()
		}
		return reports, fmt.Errorf("errors scanning packages: %s", strings.Join(errMsgs, "; "))
	}

	return reports, nil
}

// WriteConsoleReport writes vulnerability reports to console in a user-friendly format
func WriteConsoleReport(w io.Writer, reports map[string]*VulnerabilityReport) {
	// Group packages by vulnerability status
	var vulnerable []string
	var safe []string

	for pkg, report := range reports {
		if report.HasVulnerabilities() {
			vulnerable = append(vulnerable, pkg)
		} else {
			safe = append(safe, pkg)
		}
	}

	// Sort package names
	sort.Strings(vulnerable)
	sort.Strings(safe)

	// Print summary
	fmt.Fprintf(w, "%s\n", style.Title("📊 Vulnerability Scan Summary"))
	fmt.Fprintf(w, "%s\n", style.Divider())
	fmt.Fprintf(w, "%s %d\n", style.Info("Total packages scanned:"), len(reports))
	fmt.Fprintf(w, "%s %s\n", style.Warning("Vulnerable packages:"), style.Warning(fmt.Sprintf("%d", len(vulnerable))))
	fmt.Fprintf(w, "%s %s\n\n", style.Success("Clean packages:"), style.Success(fmt.Sprintf("%d", len(safe))))

	// Print vulnerable packages
	if len(vulnerable) > 0 {
		fmt.Fprintf(w, "%s\n", style.Error("🚨 Vulnerable Packages:"))
		fmt.Fprintf(w, "%s\n", style.Divider())
		for _, pkg := range vulnerable {
			report := reports[pkg]
			counts := report.CountBySeverity()

			totalVulns := len(report.Vulnerabilities)
			fmt.Fprintf(w, "%s %s: %s\n",
				style.Bullet("•"),
				style.FormatPackage(pkg, report.Version),
				style.Warning(fmt.Sprintf("%d vulnerabilities", totalVulns)))

			// Print severity counts in a nice format
			var sevCounts []string
			severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
			for _, sev := range severities {
				if count, ok := counts[sev]; ok && count > 0 {
					sevCounts = append(sevCounts, fmt.Sprintf("%d %s", count, style.GetSeverityColor(sev)))
				}
			}
			if len(sevCounts) > 0 {
				fmt.Fprintf(w, "  (%s)\n", strings.Join(sevCounts, ", "))
			}
		}
		fmt.Fprintln(w)
	}

	// Print safe packages
	if len(safe) > 0 {
		fmt.Fprintf(w, "%s\n", style.Success("✅ Clean Packages:"))
		fmt.Fprintf(w, "%s\n", style.Divider())
		for i, pkg := range safe {
			if i > 0 && i%5 == 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s %s  ",
				style.Bullet("•"),
				style.FormatPackage(pkg, reports[pkg].Version))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w)
	}

	// Print detailed reports for vulnerable packages
	if len(vulnerable) > 0 {
		fmt.Fprintf(w, "%s\n", style.Header("📝 Detailed Vulnerability Reports"))
		fmt.Fprintf(w, "%s\n\n", style.Divider())

		for _, pkg := range vulnerable {
			fmt.Fprintln(w, strings.Repeat("-", 80))
			fmt.Fprintf(w, "%s for %s\n",
				style.Subheader("Security Vulnerability Report"),
				style.FormatPackage(pkg, reports[pkg].Version))
			fmt.Fprintf(w, "Generated: %s\n\n", style.Info(reports[pkg].Timestamp.Format(time.RFC1123)))

			// Print summary
			fmt.Fprintf(w, "Found %s:\n", style.Warning(fmt.Sprintf("%d vulnerabilities", len(reports[pkg].Vulnerabilities))))
			counts := reports[pkg].CountBySeverity()
			severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}
			for _, sev := range severities {
				if count, ok := counts[sev]; ok && count > 0 {
					fmt.Fprintf(w, "  %s %s: %d\n", style.Bullet("•"), style.GetSeverityColor(sev), count)
				}
			}
			fmt.Fprintln(w)

			// Print detailed vulnerability information
			fmt.Fprintf(w, "%s:\n", style.Subheader("Vulnerability Details"))
			fmt.Fprintln(w, style.Divider())

			for i, vuln := range reports[pkg].Vulnerabilities {
				fmt.Fprintf(w, "[%d] %s (%s)\n",
					i+1,
					style.Highlight(vuln.ID),
					style.GetSeverityColor(vuln.Severity))
				fmt.Fprintf(w, "    %s %s\n", style.Info("Description:"), vuln.Description)
				fmt.Fprintf(w, "    %s %s\n", style.Info("Published:"), vuln.Published.Format("2006-01-02"))
				if len(vuln.References) > 0 {
					fmt.Fprintf(w, "    %s\n", style.Info("References:"))
					for _, ref := range vuln.References {
						fmt.Fprintf(w, "      %s %s\n", style.Bullet("•"), style.URL(ref))
					}
				}
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w)
		}
	}
}
