package security

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestVulnerabilityReport(t *testing.T) {
	// Create test vulnerabilities
	vulns := []Vulnerability{
		{
			ID:          "CVE-2023-1234",
			Package:     "test-pkg",
			Version:     "1.0.0",
			Description: "Critical security vulnerability",
			Severity:    "CRITICAL",
			Published:   time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC),
			References: []string{
				"https://example.com/cve-2023-1234",
				"https://github.com/example/test-pkg/issues/1",
			},
		},
		{
			ID:          "CVE-2023-5678",
			Package:     "test-pkg",
			Version:     "1.0.0",
			Description: "Medium severity issue",
			Severity:    "MEDIUM",
			Published:   time.Date(2023, 5, 20, 0, 0, 0, 0, time.UTC),
			References: []string{
				"https://example.com/cve-2023-5678",
			},
		},
	}

	// Create a report
	report := NewVulnerabilityReport("test-pkg", "1.0.0", vulns)

	// Test HasVulnerabilities
	t.Run("HasVulnerabilities", func(t *testing.T) {
		if !report.HasVulnerabilities() {
			t.Error("Expected report to have vulnerabilities")
		}

		emptyReport := NewVulnerabilityReport("empty-pkg", "1.0.0", nil)
		if emptyReport.HasVulnerabilities() {
			t.Error("Expected empty report to have no vulnerabilities")
		}
	})

	// Test CountBySeverity
	t.Run("CountBySeverity", func(t *testing.T) {
		counts := report.CountBySeverity()
		if counts["CRITICAL"] != 1 {
			t.Errorf("Expected 1 CRITICAL vulnerability, got %d", counts["CRITICAL"])
		}
		if counts["MEDIUM"] != 1 {
			t.Errorf("Expected 1 MEDIUM vulnerability, got %d", counts["MEDIUM"])
		}
		if counts["LOW"] != 0 {
			t.Errorf("Expected 0 LOW vulnerabilities, got %d", counts["LOW"])
		}
	})

	// Test SortBySeverity
	t.Run("SortBySeverity", func(t *testing.T) {
		// Add a HIGH vulnerability to test sorting
		highVuln := Vulnerability{
			ID:          "CVE-2023-9012",
			Package:     "test-pkg",
			Version:     "1.0.0",
			Description: "High severity issue",
			Severity:    "HIGH",
			Published:   time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
		}

		reportWithHigh := NewVulnerabilityReport("test-pkg", "1.0.0", append(vulns, highVuln))
		reportWithHigh.SortBySeverity()

		if reportWithHigh.Vulnerabilities[0].Severity != "CRITICAL" {
			t.Errorf("Expected first vulnerability to be CRITICAL, got %s", reportWithHigh.Vulnerabilities[0].Severity)
		}
		if reportWithHigh.Vulnerabilities[1].Severity != "HIGH" {
			t.Errorf("Expected second vulnerability to be HIGH, got %s", reportWithHigh.Vulnerabilities[1].Severity)
		}
		if reportWithHigh.Vulnerabilities[2].Severity != "MEDIUM" {
			t.Errorf("Expected third vulnerability to be MEDIUM, got %s", reportWithHigh.Vulnerabilities[2].Severity)
		}
	})

	// Test WriteText
	t.Run("WriteText", func(t *testing.T) {
		var buf bytes.Buffer
		err := report.WriteText(&buf)
		if err != nil {
			t.Fatalf("WriteText() error = %v", err)
		}

		output := buf.String()
		// Check if key information is present
		if !strings.Contains(output, "test-pkg@1.0.0") {
			t.Error("Expected output to contain package name and version")
		}
		if !strings.Contains(output, "CVE-2023-1234") {
			t.Error("Expected output to contain CVE ID")
		}
		if !strings.Contains(output, "CRITICAL") {
			t.Error("Expected output to contain severity level")
		}
		if !strings.Contains(output, "Critical security vulnerability") {
			t.Error("Expected output to contain vulnerability description")
		}
		if !strings.Contains(output, "https://example.com/cve-2023-1234") {
			t.Error("Expected output to contain reference URLs")
		}
	})

	// Test WriteTable
	t.Run("WriteTable", func(t *testing.T) {
		var buf bytes.Buffer
		err := report.WriteTable(&buf)
		if err != nil {
			t.Fatalf("WriteTable() error = %v", err)
		}

		output := buf.String()
		// Check if key information is present in tabular format
		if !strings.Contains(output, "ID") && !strings.Contains(output, "Severity") && !strings.Contains(output, "Published") {
			t.Error("Expected output to contain table headers")
		}
		if !strings.Contains(output, "CVE-2023-1234") {
			t.Error("Expected output to contain CVE ID")
		}
		if !strings.Contains(output, "CRITICAL") {
			t.Error("Expected output to contain severity level")
		}
	})
}

func TestVulnerabilityReporter(t *testing.T) {
	// Create a mock scanner
	mockScanner := &MockVulnerabilityScanner{
		packages: map[string][]Vulnerability{
			"vuln-pkg@1.0.0": {
				{
					ID:          "CVE-2023-1234",
					Package:     "vuln-pkg",
					Version:     "1.0.0",
					Description: "Critical security vulnerability",
					Severity:    "CRITICAL",
					Published:   time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC),
				},
			},
			"safe-pkg@2.0.0": {},
		},
	}

	reporter := NewVulnerabilityReporter(mockScanner)

	// Test GenerateReport for a package with vulnerabilities
	t.Run("GenerateReport with vulnerabilities", func(t *testing.T) {
		report, err := reporter.GenerateReport("vuln-pkg", "1.0.0")
		if err != nil {
			t.Fatalf("GenerateReport() error = %v", err)
		}

		if !report.HasVulnerabilities() {
			t.Error("Expected report to have vulnerabilities")
		}

		if len(report.Vulnerabilities) != 1 {
			t.Errorf("Expected 1 vulnerability, got %d", len(report.Vulnerabilities))
		}

		if report.Vulnerabilities[0].ID != "CVE-2023-1234" {
			t.Errorf("Expected vulnerability ID CVE-2023-1234, got %s", report.Vulnerabilities[0].ID)
		}
	})

	// Test GenerateReport for a package without vulnerabilities
	t.Run("GenerateReport without vulnerabilities", func(t *testing.T) {
		report, err := reporter.GenerateReport("safe-pkg", "2.0.0")
		if err != nil {
			t.Fatalf("GenerateReport() error = %v", err)
		}

		if report.HasVulnerabilities() {
			t.Error("Expected report to have no vulnerabilities")
		}
	})

	// Test ReportMultiple
	t.Run("ReportMultiple", func(t *testing.T) {
		packages := map[string]string{
			"vuln-pkg": "1.0.0",
			"safe-pkg": "2.0.0",
		}

		reports, err := reporter.ReportMultiple(packages)
		if err != nil {
			t.Fatalf("ReportMultiple() error = %v", err)
		}

		if len(reports) != 2 {
			t.Errorf("Expected 2 reports, got %d", len(reports))
		}

		vulnReport, ok := reports["vuln-pkg"]
		if !ok {
			t.Error("Expected to find report for vuln-pkg")
		} else if !vulnReport.HasVulnerabilities() {
			t.Error("Expected vuln-pkg report to have vulnerabilities")
		}

		safeReport, ok := reports["safe-pkg"]
		if !ok {
			t.Error("Expected to find report for safe-pkg")
		} else if safeReport.HasVulnerabilities() {
			t.Error("Expected safe-pkg report to have no vulnerabilities")
		}
	})

	// Test WriteConsoleReport
	t.Run("WriteConsoleReport", func(t *testing.T) {
		packages := map[string]string{
			"vuln-pkg": "1.0.0",
			"safe-pkg": "2.0.0",
		}

		reports, _ := reporter.ReportMultiple(packages)

		var buf bytes.Buffer
		WriteConsoleReport(&buf, reports)

		output := buf.String()

		// Check if key sections are present
		if !strings.Contains(output, "Vulnerability Scan Summary") {
			t.Error("Expected output to contain summary section")
		}

		if !strings.Contains(output, "Vulnerable Packages") {
			t.Error("Expected output to contain vulnerable packages section")
		}

		if !strings.Contains(output, "Clean Packages") {
			t.Error("Expected output to contain clean packages section")
		}

		if !strings.Contains(output, "vuln-pkg@1.0.0") {
			t.Error("Expected output to contain vulnerable package name")
		}

		if !strings.Contains(output, "safe-pkg@2.0.0") {
			t.Error("Expected output to contain safe package name")
		}

		if !strings.Contains(output, "Detailed Vulnerability Reports") {
			t.Error("Expected output to contain detailed reports section")
		}
	})
}

// MockVulnerabilityScanner implements VulnerabilityScanner for testing
type MockVulnerabilityScanner struct {
	packages map[string][]Vulnerability
}

// ScanPackage implements the VulnerabilityScanner interface for testing
func (m *MockVulnerabilityScanner) ScanPackage(name, version string) ([]Vulnerability, error) {
	key := name + "@" + version
	return m.packages[key], nil
}
