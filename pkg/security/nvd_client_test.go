package security

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HibiZA/dai/pkg/semver"
)

func TestNVDFindVulnerabilities(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request parameters
		keywordSearch := r.URL.Query().Get("keywordSearch")

		if keywordSearch != "lodash" {
			http.NotFound(w, r)
			return
		}

		// Return a sample response with vulnerabilities
		fmt.Fprint(w, `{
			"vulnerabilities": [
				{
					"cve": {
						"id": "CVE-2020-8203",
						"published": "2020-07-15T15:15:00.000Z",
						"lastModified": "2023-02-14T17:09:00.000Z",
						"descriptions": [
							{
								"lang": "en",
								"value": "Prototype pollution vulnerability in lodash before 4.17.20 allows attackers to modify object properties."
							}
						],
						"metrics": {
							"cvssMetricV31": [
								{
									"source": "nvd@nist.gov",
									"type": "Primary",
									"cvssData": {
										"version": "3.1",
										"vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:L/A:L",
										"baseScore": 7.3
									},
									"baseSeverity": "HIGH",
									"exploitabilityScore": 3.9,
									"impactScore": 3.4
								}
							]
						},
						"configurations": [
							{
								"nodes": [
									{
										"operator": "OR",
										"cpeMatch": [
											{
												"vulnerable": true,
												"criteria": "cpe:2.3:a:lodash:lodash:*:*:*:*:*:*:*:*",
												"versionEndExcluding": "4.17.20"
											}
										]
									}
								]
							}
						],
						"references": [
							{
								"url": "https://github.com/lodash/lodash/pull/4759",
								"source": "CONFIRM",
								"tags": ["Patch", "Issue Tracking"]
							},
							{
								"url": "https://hackerone.com/reports/712065",
								"source": "MISC",
								"tags": ["Third Party Advisory"]
							}
						]
					}
				}
			],
			"totalResults": 1
		}`)
	}))
	defer server.Close()

	// Create a client that uses the mock server
	client := NewNVDClient("")
	client.URL = server.URL

	// Test with a vulnerable version
	t.Run("vulnerable version", func(t *testing.T) {
		vulns, err := client.FindVulnerabilities("npm", "lodash", "4.17.19")
		if err != nil {
			t.Fatalf("FindVulnerabilities() error = %v", err)
		}

		if len(vulns) != 1 {
			t.Errorf("Expected 1 vulnerability, got %d", len(vulns))
			return
		}

		if vulns[0].ID != "CVE-2020-8203" {
			t.Errorf("Expected vulnerability ID CVE-2020-8203, got %s", vulns[0].ID)
		}

		if vulns[0].Severity != "HIGH" {
			t.Errorf("Expected severity HIGH, got %s", vulns[0].Severity)
		}

		if len(vulns[0].References) != 2 {
			t.Errorf("Expected 2 references, got %d", len(vulns[0].References))
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

func TestIsVersionInVulnerableRange(t *testing.T) {
	testCases := []struct {
		version  string
		cpeMatch CPEMatch
		expected bool
	}{
		// Testing versionEndExcluding
		{"1.2.3", CPEMatch{Vulnerable: true, VersionEndExcluding: "2.0.0"}, true},
		{"2.0.0", CPEMatch{Vulnerable: true, VersionEndExcluding: "2.0.0"}, false},

		// Testing versionStartExcluding
		{"1.2.3", CPEMatch{Vulnerable: true, VersionStartExcluding: "1.0.0"}, true},
		{"1.0.0", CPEMatch{Vulnerable: true, VersionStartExcluding: "1.0.0"}, false},

		// Testing versionEndIncluding
		{"1.2.3", CPEMatch{Vulnerable: true, VersionEndIncluding: "1.2.3"}, true},
		{"1.2.4", CPEMatch{Vulnerable: true, VersionEndIncluding: "1.2.3"}, false},

		// Testing versionStartIncluding
		{"1.2.3", CPEMatch{Vulnerable: true, VersionStartIncluding: "1.2.3"}, true},
		{"1.2.2", CPEMatch{Vulnerable: true, VersionStartIncluding: "1.2.3"}, false},

		// Testing combined ranges
		{"1.5.0", CPEMatch{Vulnerable: true, VersionStartIncluding: "1.0.0", VersionEndExcluding: "2.0.0"}, true},
		{"0.9.0", CPEMatch{Vulnerable: true, VersionStartIncluding: "1.0.0", VersionEndExcluding: "2.0.0"}, false},
		{"2.0.0", CPEMatch{Vulnerable: true, VersionStartIncluding: "1.0.0", VersionEndExcluding: "2.0.0"}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s with %+v", tc.version, tc.cpeMatch), func(t *testing.T) {
			version, err := semver.Parse(tc.version)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}

			result := isVersionInVulnerableRange(version, tc.cpeMatch)
			if result != tc.expected {
				t.Errorf("isVersionInVulnerableRange(%s, %+v) = %v, expected %v", tc.version, tc.cpeMatch, result, tc.expected)
			}
		})
	}
}

func TestGetCPEPrefix(t *testing.T) {
	testCases := []struct {
		ecosystem   string
		packageName string
		expected    string
	}{
		{"npm", "lodash", "cpe:2.3:a:lodash:lodash"},
		{"maven", "log4j-core", "cpe:2.3:a:log4j-core:log4j-core"},
		{"pypi", "django", "cpe:2.3:a:django:django"},
		{"rubygems", "rails", "cpe:2.3:a:rails:rails"},
		{"golang", "github.com/gorilla/websocket", "cpe:2.3:a:github.com/gorilla/websocket:github.com/gorilla/websocket"},
		{"unknown", "package", ""},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s:%s", tc.ecosystem, tc.packageName), func(t *testing.T) {
			result := getCPEPrefix(tc.ecosystem, tc.packageName)
			if result != tc.expected {
				t.Errorf("getCPEPrefix(%s, %s) = %s, expected %s", tc.ecosystem, tc.packageName, result, tc.expected)
			}
		})
	}
}
