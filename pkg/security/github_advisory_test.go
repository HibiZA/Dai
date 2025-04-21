package security

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HibiZA/dai/pkg/semver"
)

func TestFindVulnerabilities(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request parameters
		ecosystem := r.URL.Query().Get("ecosystem")
		pkg := r.URL.Query().Get("package")

		if ecosystem != "npm" || pkg != "lodash" {
			http.NotFound(w, r)
			return
		}

		// Return a sample response with vulnerabilities
		fmt.Fprint(w, `{
			"advisories": [
				{
					"ghsa_id": "GHSA-p6mc-m468-83gw",
					"summary": "Prototype Pollution in lodash",
					"description": "Versions of lodash prior to 4.17.20 are vulnerable to Prototype Pollution",
					"severity": "high",
					"published_at": "2020-08-17T16:27:38Z",
					"updated_at": "2020-08-17T16:27:38Z",
					"references": ["https://nvd.nist.gov/vuln/detail/CVE-2020-8203"],
					"vulnerabilities": [
						{
							"package": {
								"ecosystem": "npm",
								"name": "lodash"
							},
							"vulnerable_version_range": ["<4.17.20"],
							"first_patched_version": "4.17.20"
						}
					],
					"cvss": {
						"score": 7.3,
						"vector_string": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:L/A:L"
					},
					"cwes": [
						{
							"cwe_id": "CWE-1321",
							"name": "Prototype Pollution"
						}
					]
				}
			],
			"total_count": 1
		}`)
	}))
	defer server.Close()

	// Create a client that uses the mock server
	client := NewGitHubAdvisoryClient("")
	client.URL = server.URL

	// Test with a vulnerable version
	t.Run("vulnerable version", func(t *testing.T) {
		vulns, err := client.FindVulnerabilities("npm", "lodash", "4.17.19")
		if err != nil {
			t.Fatalf("FindVulnerabilities() error = %v", err)
		}

		if len(vulns) != 1 {
			t.Errorf("Expected 1 vulnerability, got %d", len(vulns))
		}

		if vulns[0].ID != "GHSA-p6mc-m468-83gw" {
			t.Errorf("Expected vulnerability ID GHSA-p6mc-m468-83gw, got %s", vulns[0].ID)
		}

		if vulns[0].Severity != "high" {
			t.Errorf("Expected severity high, got %s", vulns[0].Severity)
		}
	})

	// Test with a patched version
	t.Run("patched version", func(t *testing.T) {
		vulns, err := client.FindVulnerabilities("npm", "lodash", "4.17.20")
		if err != nil {
			t.Fatalf("FindVulnerabilities() error = %v", err)
		}

		if len(vulns) != 0 {
			t.Errorf("Expected 0 vulnerabilities for patched version, got %d", len(vulns))
		}
	})
}

func TestIsVersionVulnerable(t *testing.T) {
	testCases := []struct {
		version          string
		vulnerableRanges []string
		expected         bool
	}{
		{"1.2.3", []string{"1.2.3"}, true},                    // Exact match
		{"1.2.3", []string{"<2.0.0"}, true},                   // Less than
		{"2.0.0", []string{"<2.0.0"}, false},                  // Boundary
		{"1.2.3", []string{"<=1.2.3"}, true},                  // Less than or equal
		{"1.2.4", []string{"<=1.2.3"}, false},                 // Greater than range
		{"1.2.3", []string{">1.0.0"}, true},                   // Greater than
		{"1.0.0", []string{">1.0.0"}, false},                  // Boundary
		{"1.2.3", []string{">=1.0.0"}, true},                  // Greater than or equal
		{"1.2.3", []string{">=1.0.0 <2.0.0"}, true},           // Within range
		{"2.0.0", []string{">=1.0.0 <2.0.0"}, false},          // Upper boundary
		{"0.9.0", []string{">=1.0.0 <2.0.0"}, false},          // Lower boundary
		{"1.2.3", []string{"not a valid range"}, false},       // Invalid range
		{"1.2.3", []string{"<1.0.0", ">=1.2.0 <1.3.0"}, true}, // Multiple ranges, one matches
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %v", tc.version, tc.vulnerableRanges), func(t *testing.T) {
			version, err := semver.Parse(tc.version)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}

			result := isVersionVulnerable(version, tc.vulnerableRanges)
			if result != tc.expected {
				t.Errorf("isVersionVulnerable(%s, %v) = %v, expected %v", tc.version, tc.vulnerableRanges, result, tc.expected)
			}
		})
	}
}
