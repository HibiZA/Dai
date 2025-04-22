package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HibiZA/dai/pkg/semver"
)

const (
	// GitHubAdvisoryURL is the base URL for the GitHub Advisory Database API
	GitHubAdvisoryURL = "https://api.github.com/advisories"
)

// GitHubAdvisory represents a security advisory from GitHub
type GitHubAdvisory struct {
	ID              string              `json:"ghsa_id"`
	Summary         string              `json:"summary"`
	Description     string              `json:"description"`
	Severity        string              `json:"severity"`
	PublishedAt     time.Time           `json:"published_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	WithdrawnAt     time.Time           `json:"withdrawn_at,omitempty"`
	CVSS            CVSS                `json:"cvss"`
	CWEs            []CWE               `json:"cwes"`
	Vulnerabilities []VulnerabilityInfo `json:"vulnerabilities"`
	References      []string            `json:"references"`
	IdentifierURLs  []string            `json:"identifiers"`
}

// CVSS represents Common Vulnerability Scoring System data
type CVSS struct {
	Score        float64 `json:"score"`
	VectorString string  `json:"vector_string"`
}

// CWE represents a Common Weakness Enumeration
type CWE struct {
	ID   string `json:"cwe_id"`
	Name string `json:"name"`
}

// VulnerabilityInfo represents information about a vulnerable package
type VulnerabilityInfo struct {
	Package             PackageInfo `json:"package"`
	VulnerableVersions  []string    `json:"vulnerable_version_range"`
	FirstPatchedVersion string      `json:"first_patched_version,omitempty"`
}

// PackageInfo represents a package identified in a vulnerability
type PackageInfo struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

// AdvisoryResponse represents the API response from GitHub
type AdvisoryResponse struct {
	Advisories []GitHubAdvisory `json:"advisories"`
	TotalCount int              `json:"total_count"`
}

// GitHubAdvisoryClient implements security scanning using GitHub Advisory DB
type GitHubAdvisoryClient struct {
	URL        string
	Token      string
	httpClient *http.Client
}

// NewGitHubAdvisoryClient creates a new GitHub Advisory client
func NewGitHubAdvisoryClient(token string) *GitHubAdvisoryClient {
	return &GitHubAdvisoryClient{
		URL:   GitHubAdvisoryURL,
		Token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FindVulnerabilities searches for vulnerabilities for a specific package
func (c *GitHubAdvisoryClient) FindVulnerabilities(ecosystem, packageName, version string) ([]Vulnerability, error) {
	// Build the URL with query parameters
	url := fmt.Sprintf("%s?ecosystem=%s&package=%s", c.URL, ecosystem, packageName)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var response AdvisoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process the advisories to find vulnerabilities affecting the specified version
	var vulnerabilities []Vulnerability

	// Normalize and parse the version
	normalizedVersion := semver.NormalizeVersion(version)
	parsedVersion, err := semver.Parse(normalizedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	for _, advisory := range response.Advisories {
		for _, vuln := range advisory.Vulnerabilities {
			if vuln.Package.Ecosystem == ecosystem && vuln.Package.Name == packageName {
				// Check if the specified version is vulnerable
				if isVersionVulnerable(parsedVersion, vuln.VulnerableVersions) {
					vulnerability := Vulnerability{
						ID:          advisory.ID,
						Package:     packageName,
						Version:     version,
						Description: advisory.Description,
						Severity:    advisory.Severity,
						Published:   advisory.PublishedAt,
						References:  advisory.References,
					}
					vulnerabilities = append(vulnerabilities, vulnerability)
					break // Found a match for this advisory
				}
			}
		}
	}

	return vulnerabilities, nil
}

// isVersionVulnerable checks if a version is within any of the vulnerable ranges
func isVersionVulnerable(version *semver.Version, vulnerableRanges []string) bool {
	for _, rangeStr := range vulnerableRanges {
		// Handle different range formats from GitHub Advisory DB
		// For simplicity, we'll just do basic checks for common patterns

		// Check for exact version match
		if rangeStr == version.String() {
			return true
		}

		// Check for "< x.y.z" pattern
		if strings.HasPrefix(rangeStr, "<") {
			upperBoundStr := strings.TrimSpace(strings.TrimPrefix(rangeStr, "<"))
			upperBound, err := semver.Parse(upperBoundStr)
			if err == nil {
				if semver.Compare(version, upperBound) < 0 {
					return true
				}
			}
		}

		// Check for "<= x.y.z" pattern
		if strings.HasPrefix(rangeStr, "<=") {
			upperBoundStr := strings.TrimSpace(strings.TrimPrefix(rangeStr, "<="))
			upperBound, err := semver.Parse(upperBoundStr)
			if err == nil {
				if semver.Compare(version, upperBound) <= 0 {
					return true
				}
			}
		}

		// Check for >= x.y.z pattern
		if strings.HasPrefix(rangeStr, ">=") {
			lowerBoundStr := strings.TrimSpace(strings.TrimPrefix(rangeStr, ">="))
			lowerBound, err := semver.Parse(lowerBoundStr)
			if err == nil {
				if semver.Compare(version, lowerBound) >= 0 {
					return true
				}
			}
		}

		// Check for > x.y.z pattern
		if strings.HasPrefix(rangeStr, ">") && !strings.HasPrefix(rangeStr, ">=") {
			lowerBoundStr := strings.TrimSpace(strings.TrimPrefix(rangeStr, ">"))
			lowerBound, err := semver.Parse(lowerBoundStr)
			if err == nil {
				if semver.Compare(version, lowerBound) > 0 {
					return true
				}
			}
		}

		// Check for ranges like ">=1.0.0 <2.0.0"
		if strings.Contains(rangeStr, " ") {
			parts := strings.Split(rangeStr, " ")
			if len(parts) == 2 {
				// Handle lower bound
				lowerBoundValid := false
				var lowerBound *semver.Version
				var err error

				if strings.HasPrefix(parts[0], ">=") {
					lowerBoundStr := strings.TrimSpace(strings.TrimPrefix(parts[0], ">="))
					lowerBound, err = semver.Parse(lowerBoundStr)
					if err == nil {
						lowerBoundValid = semver.Compare(version, lowerBound) >= 0
					}
				} else if strings.HasPrefix(parts[0], ">") {
					lowerBoundStr := strings.TrimSpace(strings.TrimPrefix(parts[0], ">"))
					lowerBound, err = semver.Parse(lowerBoundStr)
					if err == nil {
						lowerBoundValid = semver.Compare(version, lowerBound) > 0
					}
				}

				// Handle upper bound
				upperBoundValid := false
				var upperBound *semver.Version

				if strings.HasPrefix(parts[1], "<=") {
					upperBoundStr := strings.TrimSpace(strings.TrimPrefix(parts[1], "<="))
					upperBound, err = semver.Parse(upperBoundStr)
					if err == nil {
						upperBoundValid = semver.Compare(version, upperBound) <= 0
					}
				} else if strings.HasPrefix(parts[1], "<") {
					upperBoundStr := strings.TrimSpace(strings.TrimPrefix(parts[1], "<"))
					upperBound, err = semver.Parse(upperBoundStr)
					if err == nil {
						upperBoundValid = semver.Compare(version, upperBound) < 0
					}
				}

				// If both bounds are valid, the version is vulnerable
				if lowerBoundValid && upperBoundValid {
					return true
				}
			}
		}
	}

	return false
}
