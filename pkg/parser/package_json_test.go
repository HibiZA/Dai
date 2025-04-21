package parser

import (
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
