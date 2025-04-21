package npm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPackageInfo(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/react" {
			http.NotFound(w, r)
			return
		}

		// Return a simplified mock response
		fmt.Fprint(w, `{
			"name": "react",
			"description": "React is a JavaScript library for building user interfaces.",
			"dist-tags": {
				"latest": "18.2.0"
			},
			"versions": {
				"16.14.0": {
					"name": "react",
					"version": "16.14.0",
					"description": "React is a JavaScript library for building user interfaces.",
					"dependencies": {
						"loose-envify": "^1.1.0",
						"object-assign": "^4.1.1",
						"prop-types": "^15.6.2"
					}
				},
				"17.0.2": {
					"name": "react",
					"version": "17.0.2",
					"description": "React is a JavaScript library for building user interfaces.",
					"dependencies": {
						"loose-envify": "^1.1.0",
						"object-assign": "^4.1.1"
					}
				},
				"18.2.0": {
					"name": "react",
					"version": "18.2.0",
					"description": "React is a JavaScript library for building user interfaces.",
					"dependencies": {
						"loose-envify": "^1.1.0"
					}
				}
			}
		}`)
	}))
	defer server.Close()

	// Create a client that uses the mock server
	client := NewRegistryClient(server.URL)

	// Test GetPackageInfo
	info, err := client.GetPackageInfo("react")
	if err != nil {
		t.Fatalf("GetPackageInfo() error = %v", err)
	}

	if info.Name != "react" {
		t.Errorf("Expected package name 'react', got '%s'", info.Name)
	}

	if len(info.Versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(info.Versions))
	}

	// Test GetLatestVersion
	latest, err := client.GetLatestVersion("react")
	if err != nil {
		t.Fatalf("GetLatestVersion() error = %v", err)
	}

	if latest != "18.2.0" {
		t.Errorf("Expected latest version '18.2.0', got '%s'", latest)
	}

	// Test FindBestUpgrade with caret range
	upgrade, err := client.FindBestUpgrade("react", "^17.0.1")
	if err != nil {
		t.Fatalf("FindBestUpgrade() error = %v", err)
	}

	if upgrade != "^17.0.2" {
		t.Errorf("Expected upgrade '^17.0.2', got '%s'", upgrade)
	}

	// Test FindBestUpgrade with exact version
	upgrade, err = client.FindBestUpgrade("react", "16.14.0")
	if err != nil {
		t.Fatalf("FindBestUpgrade() error = %v", err)
	}

	// Should find the latest within the major version
	if upgrade != "16.14.0" {
		t.Errorf("Expected upgrade '16.14.0', got '%s'", upgrade)
	}
}
