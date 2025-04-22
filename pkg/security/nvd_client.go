package security

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/HibiZA/dai/pkg/semver"
)

const (
	// NVDApiURL is the base URL for the NVD API
	NVDApiURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
)

// NVDClient is a client for the National Vulnerability Database API
type NVDClient struct {
	URL        string
	ApiKey     string
	httpClient *http.Client
}

// NewNVDClient creates a new NVD API client
func NewNVDClient(apiKey string) *NVDClient {
	return &NVDClient{
		URL:    NVDApiURL,
		ApiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NVDResponse represents the response from the NVD API
type NVDResponse struct {
	Vulnerabilities []NVDVulnerability `json:"vulnerabilities"`
	TotalResults    int                `json:"totalResults"`
}

// NVDVulnerability represents a vulnerability from the NVD API
type NVDVulnerability struct {
	CVE CVEItem `json:"cve"`
}

// CVEItem represents a CVE entry in the NVD database
type CVEItem struct {
	ID                 string           `json:"id"`
	Published          string           `json:"published"`
	LastModified       string           `json:"lastModified"`
	Descriptions       []CVEDescription `json:"descriptions"`
	Metrics            CVEMetrics       `json:"metrics"`
	Configurations     []CVEConfig      `json:"configurations"`
	References         []CVEReference   `json:"references"`
	VulnerableSoftware []CVESoftware    `json:"vulnerableSoftware"`
}

// CVEDescription represents a description of a CVE
type CVEDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// CVEMetrics represents CVSS metrics for a CVE
type CVEMetrics struct {
	CVSSMetricV31 []CVSS31 `json:"cvssMetricV31"`
	CVSSMetricV30 []CVSS30 `json:"cvssMetricV30"`
	CVSSMetricV2  []CVSS2  `json:"cvssMetricV2"`
}

// CVSS31 represents CVSS 3.1 metrics
type CVSS31 struct {
	Source       string     `json:"source"`
	Type         string     `json:"type"`
	CVSSData     CVSSData31 `json:"cvssData"`
	BaseSeverity string     `json:"baseSeverity"`
	ExploitScore float64    `json:"exploitabilityScore"`
	ImpactScore  float64    `json:"impactScore"`
}

// CVSSData31 contains the actual CVSS 3.1 data
type CVSSData31 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
}

// CVSS30 represents CVSS 3.0 metrics
type CVSS30 struct {
	Source       string     `json:"source"`
	Type         string     `json:"type"`
	CVSSData     CVSSData30 `json:"cvssData"`
	BaseSeverity string     `json:"baseSeverity"`
	ExploitScore float64    `json:"exploitabilityScore"`
	ImpactScore  float64    `json:"impactScore"`
}

// CVSSData30 contains the actual CVSS 3.0 data
type CVSSData30 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
}

// CVSS2 represents CVSS 2.0 metrics
type CVSS2 struct {
	Source       string    `json:"source"`
	Type         string    `json:"type"`
	CVSSData     CVSSData2 `json:"cvssData"`
	ExploitScore float64   `json:"exploitabilityScore"`
	ImpactScore  float64   `json:"impactScore"`
}

// CVSSData2 contains the actual CVSS 2.0 data
type CVSSData2 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
}

// CVEConfig represents a configuration for a vulnerable product
type CVEConfig struct {
	Nodes []CVENode `json:"nodes"`
}

// CVENode represents a node in the configuration tree
type CVENode struct {
	Operator string     `json:"operator"`
	CPEMatch []CPEMatch `json:"cpeMatch"`
	Children []CVENode  `json:"children"`
}

// CPEMatch represents a CPE match condition
type CPEMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	CPE23URI              string `json:"criteria"`
	VersionStartExcluding string `json:"versionStartExcluding"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
	VersionEndIncluding   string `json:"versionEndIncluding"`
}

// CVEReference represents a reference for a CVE
type CVEReference struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}

// CVESoftware represents vulnerable software
type CVESoftware struct {
	CPE23URI string `json:"criteria"`
}

// FindVulnerabilities searches for vulnerabilities for a specific package
func (c *NVDClient) FindVulnerabilities(ecosystem, packageName, version string) ([]Vulnerability, error) {
	// Convert ecosystem and package name to CPE format for filtering
	cpePrefix := getCPEPrefix(ecosystem, packageName)
	if cpePrefix == "" {
		return nil, fmt.Errorf("unsupported ecosystem: %s", ecosystem)
	}

	// Normalize the version before processing
	normalizedVersion := semver.NormalizeVersion(version)

	// Build the query URL
	queryParams := url.Values{}
	queryParams.Add("keywordSearch", packageName)

	// Set the API key if available
	req, err := http.NewRequest("GET", c.URL+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	if c.ApiKey != "" {
		req.Header.Set("apiKey", c.ApiKey)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NVD API returned status %d", resp.StatusCode)
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var response NVDResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Process the vulnerabilities to find those affecting the specified version
	var vulnerabilities []Vulnerability

	// Parse the normalized version
	parsedVersion, err := semver.Parse(normalizedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	for _, nvdVuln := range response.Vulnerabilities {
		// Filter by CPE that matches our package and ecosystem
		isVulnerable := false
		for _, config := range nvdVuln.CVE.Configurations {
			for _, node := range config.Nodes {
				for _, cpeMatch := range node.CPEMatch {
					if cpeMatch.Vulnerable && strings.Contains(cpeMatch.CPE23URI, cpePrefix) {
						// Check if the specified version is in the vulnerable range
						if isVersionInVulnerableRange(parsedVersion, cpeMatch) {
							isVulnerable = true
							break
						}
					}
				}
				if isVulnerable {
					break
				}
			}
			if isVulnerable {
				break
			}
		}

		if isVulnerable {
			// Get the description (English preferred)
			var description string
			for _, desc := range nvdVuln.CVE.Descriptions {
				if desc.Lang == "en" {
					description = desc.Value
					break
				}
			}

			// Get severity
			severity := "unknown"
			if len(nvdVuln.CVE.Metrics.CVSSMetricV31) > 0 {
				severity = nvdVuln.CVE.Metrics.CVSSMetricV31[0].BaseSeverity
			} else if len(nvdVuln.CVE.Metrics.CVSSMetricV30) > 0 {
				severity = nvdVuln.CVE.Metrics.CVSSMetricV30[0].BaseSeverity
			} else if len(nvdVuln.CVE.Metrics.CVSSMetricV2) > 0 {
				// Convert CVSS 2.0 score to severity
				score := nvdVuln.CVE.Metrics.CVSSMetricV2[0].CVSSData.BaseScore
				if score >= 7.0 {
					severity = "HIGH"
				} else if score >= 4.0 {
					severity = "MEDIUM"
				} else {
					severity = "LOW"
				}
			}

			// Get references
			var references []string
			for _, ref := range nvdVuln.CVE.References {
				references = append(references, ref.URL)
			}

			// Parse published date
			published, _ := time.Parse(time.RFC3339, nvdVuln.CVE.Published)

			vulnerability := Vulnerability{
				ID:          nvdVuln.CVE.ID,
				Package:     packageName,
				Version:     version,
				Description: description,
				Severity:    severity,
				Published:   published,
				References:  references,
			}

			vulnerabilities = append(vulnerabilities, vulnerability)
		}
	}

	return vulnerabilities, nil
}

// getCPEPrefix returns the CPE URI prefix for the given ecosystem and package
func getCPEPrefix(ecosystem, packageName string) string {
	switch ecosystem {
	case "npm":
		return fmt.Sprintf("cpe:2.3:a:%s:%s", packageName, packageName)
	case "maven":
		return fmt.Sprintf("cpe:2.3:a:%s:%s", packageName, packageName)
	case "pypi":
		return fmt.Sprintf("cpe:2.3:a:%s:%s", packageName, packageName)
	case "rubygems":
		return fmt.Sprintf("cpe:2.3:a:%s:%s", packageName, packageName)
	case "golang":
		return fmt.Sprintf("cpe:2.3:a:%s:%s", packageName, packageName)
	default:
		return ""
	}
}

// isVersionInVulnerableRange checks if the version is within the vulnerable range
func isVersionInVulnerableRange(version *semver.Version, cpeMatch CPEMatch) bool {
	// Handle version ranges from CPE match
	if cpeMatch.VersionStartExcluding != "" {
		startVersion, err := semver.Parse(cpeMatch.VersionStartExcluding)
		if err == nil && semver.Compare(version, startVersion) <= 0 {
			return false
		}
	}

	if cpeMatch.VersionStartIncluding != "" {
		startVersion, err := semver.Parse(cpeMatch.VersionStartIncluding)
		if err == nil && semver.Compare(version, startVersion) < 0 {
			return false
		}
	}

	if cpeMatch.VersionEndExcluding != "" {
		endVersion, err := semver.Parse(cpeMatch.VersionEndExcluding)
		if err == nil && semver.Compare(version, endVersion) >= 0 {
			return false
		}
	}

	if cpeMatch.VersionEndIncluding != "" {
		endVersion, err := semver.Parse(cpeMatch.VersionEndIncluding)
		if err == nil && semver.Compare(version, endVersion) > 0 {
			return false
		}
	}

	// If no specific version constraints, and the CPE URI has our package, it's vulnerable
	return true
}
