package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePackageJSON(t *testing.T) {
	// Path to test package.json
	testPath := filepath.Join("..", "..", "test", "sample")

	pkg, err := ParsePackageJSON(testPath)
	if err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	// Check basic properties
	if pkg.Name != "dai-test-project" {
		t.Errorf("Expected name to be 'dai-test-project', got '%s'", pkg.Name)
	}

	if pkg.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", pkg.Version)
	}

	// Check dependencies
	expectedDeps := map[string]string{
		"express": "^4.17.1",
		"react":   "^17.0.2",
		"lodash":  "^4.17.21",
		"axios":   "^0.21.1",
	}

	for dep, version := range expectedDeps {
		if pkg.Dependencies[dep] != version {
			t.Errorf("Expected dependency %s to be %s, got %s", dep, version, pkg.Dependencies[dep])
		}
	}

	// Check dev dependencies
	expectedDevDeps := map[string]string{
		"jest":       "^27.0.6",
		"eslint":     "^7.32.0",
		"typescript": "^4.3.5",
	}

	for dep, version := range expectedDevDeps {
		if pkg.DevDependencies[dep] != version {
			t.Errorf("Expected devDependency %s to be %s, got %s", dep, version, pkg.DevDependencies[dep])
		}
	}
}

func TestUpdateDependency(t *testing.T) {
	pkg := &PackageJSON{
		Name:    "test-project",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"react": "^17.0.2",
		},
		DevDependencies: map[string]string{
			"jest": "^27.0.6",
		},
	}

	// Test updating a regular dependency
	updated := pkg.UpdateDependency("react", "^18.0.0")
	if !updated {
		t.Errorf("UpdateDependency() returned false for existing dependency")
	}
	if pkg.Dependencies["react"] != "^18.0.0" {
		t.Errorf("Expected dependency version to be ^18.0.0, got %s", pkg.Dependencies["react"])
	}

	// Test updating a dev dependency
	updated = pkg.UpdateDependency("jest", "^28.0.0")
	if !updated {
		t.Errorf("UpdateDependency() returned false for existing dev dependency")
	}
	if pkg.DevDependencies["jest"] != "^28.0.0" {
		t.Errorf("Expected dev dependency version to be ^28.0.0, got %s", pkg.DevDependencies["jest"])
	}

	// Test updating a non-existent dependency
	updated = pkg.UpdateDependency("nonexistent", "^1.0.0")
	if updated {
		t.Errorf("UpdateDependency() returned true for non-existent dependency")
	}
}

func TestWriteToFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pkg-json-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a package.json object
	pkg := &PackageJSON{
		Name:    "test-project",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"react": "^17.0.2",
		},
		DevDependencies: map[string]string{
			"jest": "^27.0.6",
		},
	}

	// Write it to the temp directory
	err = pkg.WriteToFile(tmpDir)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// Read it back
	readPkg, err := ParsePackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse written package.json: %v", err)
	}

	// Verify contents
	if readPkg.Name != pkg.Name {
		t.Errorf("Expected name to be %s, got %s", pkg.Name, readPkg.Name)
	}
	if readPkg.Version != pkg.Version {
		t.Errorf("Expected version to be %s, got %s", pkg.Version, readPkg.Version)
	}
	if readPkg.Dependencies["react"] != pkg.Dependencies["react"] {
		t.Errorf("Expected react version to be %s, got %s", pkg.Dependencies["react"], readPkg.Dependencies["react"])
	}
	if readPkg.DevDependencies["jest"] != pkg.DevDependencies["jest"] {
		t.Errorf("Expected jest version to be %s, got %s", pkg.DevDependencies["jest"], readPkg.DevDependencies["jest"])
	}
}
