package npm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/dai/pkg/semver"
)

const (
	// DefaultRegistry is the default npm registry URL
	DefaultRegistry = "https://registry.npmjs.org"
)

// RegistryClient handles interactions with the npm registry
type RegistryClient struct {
	RegistryURL string
	httpClient  *http.Client
}

// PackageInfo represents the metadata for a package from the npm registry
type PackageInfo struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Versions    map[string]PackageVersion `json:"versions"`
	Time        map[string]string         `json:"time"`
	DistTags    map[string]string         `json:"dist-tags"`
}

// PackageVersion represents a single version of a package
type PackageVersion struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
	Dist            DistInfo          `json:"dist"`
}

// DistInfo contains distribution information for a package version
type DistInfo struct {
	Shasum  string `json:"shasum"`
	Tarball string `json:"tarball"`
}

// NewRegistryClient creates a new npm registry client
func NewRegistryClient(registryURL string) *RegistryClient {
	if registryURL == "" {
		registryURL = DefaultRegistry
	}

	return &RegistryClient{
		RegistryURL: registryURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPackageInfo retrieves package information from the registry
func (c *RegistryClient) GetPackageInfo(packageName string) (*PackageInfo, error) {
	url := fmt.Sprintf("%s/%s", c.RegistryURL, packageName)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry returned status %d", resp.StatusCode)
	}

	var packageInfo PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode package info: %w", err)
	}

	return &packageInfo, nil
}

// GetLatestVersion gets the latest version of a package
func (c *RegistryClient) GetLatestVersion(packageName string) (string, error) {
	info, err := c.GetPackageInfo(packageName)
	if err != nil {
		return "", err
	}

	latest, ok := info.DistTags["latest"]
	if !ok {
		return "", fmt.Errorf("latest version not found for %s", packageName)
	}

	return latest, nil
}

// FindBestUpgrade finds the best upgrade for a package given the current version
// It respects semver constraints like ^, ~, etc.
func (c *RegistryClient) FindBestUpgrade(packageName, currentVersion string) (string, error) {
	// Parse the constraint
	constraintType, version, err := semver.ParseConstraint(currentVersion)
	if err != nil {
		return "", err
	}

	// Get package info from registry
	info, err := c.GetPackageInfo(packageName)
	if err != nil {
		return "", err
	}

	// Find the best upgrade based on the constraint type
	var bestVersion *semver.Version
	var bestVersionStr string

	for versionStr := range info.Versions {
		candidateVersion, err := semver.Parse(versionStr)
		if err != nil {
			// Skip versions that don't parse
			continue
		}

		// Check if this version is an eligible upgrade
		eligible := false

		switch constraintType {
		case "^":
			// Caret allows changes that do not modify the first non-zero digit
			if version.Major > 0 {
				eligible = candidateVersion.Major == version.Major && semver.Compare(candidateVersion, version) > 0
			} else if version.Minor > 0 {
				eligible = candidateVersion.Major == 0 && candidateVersion.Minor == version.Minor && semver.Compare(candidateVersion, version) > 0
			} else {
				eligible = candidateVersion.Major == 0 && candidateVersion.Minor == 0 && semver.Compare(candidateVersion, version) > 0
			}
		case "~":
			// Tilde allows patch-level changes if a minor version is specified
			eligible = candidateVersion.Major == version.Major && candidateVersion.Minor == version.Minor && semver.Compare(candidateVersion, version) > 0
		case ">":
			eligible = semver.Compare(candidateVersion, version) > 0
		case ">=":
			eligible = semver.Compare(candidateVersion, version) >= 0
		case "<":
			eligible = semver.Compare(candidateVersion, version) < 0
		case "<=":
			eligible = semver.Compare(candidateVersion, version) <= 0
		case "=", "":
			// For exact version match, prefer increasing the minor version within the same major
			eligible = candidateVersion.Major == version.Major && semver.Compare(candidateVersion, version) > 0
		}

		if eligible && (bestVersion == nil || semver.Compare(candidateVersion, bestVersion) > 0) {
			bestVersion = candidateVersion
			bestVersionStr = versionStr
		}
	}

	// If no better version found and it's an exact version, return the same version
	if bestVersion == nil && (constraintType == "=" || constraintType == "") {
		// Check if the version exists in the registry
		if _, exists := info.Versions[version.String()]; exists {
			return version.String(), nil
		}
		// Otherwise, try to find the exact string match in the registry
		for versionStr := range info.Versions {
			if versionStr == currentVersion || versionStr == version.String() {
				return versionStr, nil
			}
		}
	}

	if bestVersion == nil {
		return "", fmt.Errorf("no compatible upgrades found for %s@%s", packageName, currentVersion)
	}

	// Format with the original constraint
	return constraintType + bestVersionStr, nil
}
