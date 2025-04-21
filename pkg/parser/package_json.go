package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PackageJSON represents the structure of a package.json file
type PackageJSON struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Dependencies    map[string]string      `json:"dependencies"`
	DevDependencies map[string]string      `json:"devDependencies"`
	Scripts         map[string]string      `json:"scripts"`
	Other           map[string]interface{} `json:"-"`
}

// ParsePackageJSON parses a package.json file in the given directory
func ParsePackageJSON(dir string) (*PackageJSON, error) {
	packagePath := filepath.Join(dir, "package.json")

	content, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return &pkg, nil
}

// FindPackageJSON looks for a package.json file in the current directory and its parents
func FindPackageJSON() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		packagePath := filepath.Join(dir, "package.json")
		if _, err := os.Stat(packagePath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("package.json not found in current directory or any parent directory")
}

// UpdateDependency updates a dependency version in package.json
func (p *PackageJSON) UpdateDependency(name, version string) bool {
	// Check if it's a regular dependency
	if _, exists := p.Dependencies[name]; exists {
		p.Dependencies[name] = version
		return true
	}

	// Check if it's a dev dependency
	if _, exists := p.DevDependencies[name]; exists {
		p.DevDependencies[name] = version
		return true
	}

	return false
}

// WriteToFile writes the package.json content back to a file
func (p *PackageJSON) WriteToFile(dir string) error {
	packagePath := filepath.Join(dir, "package.json")

	// Format the JSON with indentation for readability
	data, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}

	// Write the file
	if err := os.WriteFile(packagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	return nil
}
