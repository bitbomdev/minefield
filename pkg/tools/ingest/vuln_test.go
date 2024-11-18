package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitbomdev/minefield/pkg/graph"
)

func TestVulnerabilities(t *testing.T) {
	storage := graph.NewMockStorage()

	vulnsDir := "../../../testdata/osv-vulns"
	sbomDir := "../../../testdata/osv-sboms"

	// Read SBOM files
	sbomFiles, err := os.ReadDir(sbomDir)
	if err != nil {
		t.Fatalf("Failed to read SBOM directory: %v", err)
	}

	for _, file := range sbomFiles {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sbomDir, file.Name()))
		if err != nil {
			t.Fatalf("Failed to read SBOM file %s: %v", file.Name(), err)
		}

		if err := SBOM(storage, data); err != nil {
			t.Fatalf("Failed to load SBOM from file %s: %v", file.Name(), err)
		}
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	numberOfNodes := len(keys)

	vulnFiles, err := os.ReadDir(vulnsDir)
	if err != nil {
		t.Fatalf("Failed to read vulnerabilities directory: %v", err)
	}

	vulnCount := 0
	for _, file := range vulnFiles {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(vulnsDir, file.Name()))
		if err != nil {
			t.Fatalf("Failed to read vulnerability file %s: %v", file.Name(), err)
		}

		if err := Vulnerabilities(storage, data); err != nil {
			t.Fatalf("Failed to load vulnerabilities from file %s: %v", file.Name(), err)
		}
		vulnCount++
	}

	if vulnCount == 0 {
		t.Fatal("Expected vulnerabilities to be ingested, got 0")
	}

	keys, err = storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	if len(keys) != numberOfNodes+3 {
		t.Fatalf("Expected number of nodes to be %d, got %d", numberOfNodes+3, len(keys))
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name      string
		v1        string
		v2        string
		eventType string
		ecosystem string
		want      int
	}{
		// SEMVER tests
		{
			name:      "SEMVER - v1 less than v2",
			v1:        "1.0.0",
			v2:        "2.0.0",
			eventType: "SEMVER",
			want:      -1,
		},
		{
			name:      "SEMVER - v1 equal to v2",
			v1:        "1.0.0",
			v2:        "1.0.0",
			eventType: "SEMVER",
			want:      0,
		},
		{
			name:      "SEMVER - v1 greater than v2",
			v1:        "2.0.0",
			v2:        "1.0.0",
			eventType: "SEMVER",
			want:      1,
		},
		{
			name:      "SEMVER - both invalid",
			v1:        "invalid1",
			v2:        "invalid2",
			eventType: "SEMVER",
			want:      -1, // falls back to string comparison
		},

		// ECOSYSTEM tests
		{
			name:      "ECOSYSTEM - v1 less than v2",
			v1:        "1.0.0",
			v2:        "2.0.0",
			eventType: "ECOSYSTEM",
			ecosystem: "npm",
			want:      -1,
		},
		{
			name:      "ECOSYSTEM - v1 equal to v2",
			v1:        "1.0.0",
			v2:        "1.0.0",
			eventType: "ECOSYSTEM",
			ecosystem: "npm",
			want:      0,
		},
		{
			name:      "ECOSYSTEM - v1 greater than v2",
			v1:        "2.0.0",
			v2:        "1.0.0",
			eventType: "ECOSYSTEM",
			ecosystem: "npm",
			want:      1,
		},

		// GIT tests
		{
			name:      "GIT - v1 less than v2",
			v1:        "abc",
			v2:        "def",
			eventType: "GIT",
			want:      -1,
		},
		{
			name:      "GIT - v1 equal to v2",
			v1:        "abc",
			v2:        "abc",
			eventType: "GIT",
			want:      0,
		},
		{
			name:      "GIT - v1 greater than v2",
			v1:        "def",
			v2:        "abc",
			eventType: "GIT",
			want:      1,
		},

		// Default case
		{
			name:      "UNKNOWN - fallback to string comparison",
			v1:        "1.0.0",
			v2:        "2.0.0",
			eventType: "UNKNOWN",
			want:      -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2, tt.eventType, tt.ecosystem)
			if got != tt.want {
				t.Errorf("compareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareEcosystemVersions(t *testing.T) {
	tests := []struct {
		name      string
		v1        string
		v2        string
		ecosystem string
		want      int
	}{
		{
			name:      "v1 less than v2",
			v1:        "1.0.0",
			v2:        "2.0.0",
			ecosystem: "npm",
			want:      -1,
		},
		{
			name:      "v1 equal to v2",
			v1:        "1.0.0",
			v2:        "1.0.0",
			ecosystem: "npm",
			want:      0,
		},
		{
			name:      "v1 greater than v2",
			v1:        "2.0.0",
			v2:        "1.0.0",
			ecosystem: "npm",
			want:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareEcosystemVersions(tt.v1, tt.v2, tt.ecosystem)
			if got != tt.want {
				t.Errorf("compareEcosystemVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
