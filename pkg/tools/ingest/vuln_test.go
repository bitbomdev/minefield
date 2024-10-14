package ingest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
)

func TestVulnerabilitiesToStorage(t *testing.T) {
	storage := graph.NewMockStorage()
	vulnsDir := "../../../testdata/vulns"
	entries, err := os.ReadDir(vulnsDir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", vulnsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		filePath := filepath.Join(vulnsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", filePath, err)
		}
		if err := LoadVulnerabilities(storage, content); err != nil {
			t.Fatalf("Failed to load vulnerabilities from file %s: %v", filePath, err)
		}
	}
	// Verify data in storage
	data, err := storage.GetCustomData(OSVTagName, "setuptools")
	if err != nil {
		t.Fatalf("Failed to get data from storage: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("Expected vulnerabilities in storage, got %d", len(data))
	}
}

func TestVulnerabilities(t *testing.T) {
	storage := graph.NewMockStorage()

	vulnsDir := "../../../testdata/osv-vulns"

	count, err := SBOM("../../../testdata/osv-sboms", storage, nil)
	if err != nil {
		t.Fatalf("Failed to ingest SBOM: %v", err)
	}
	if count == 0 {
		t.Fatalf("Expected SBOM to be ingested, got %d", count)
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	numberOfNodes := len(keys)
	entries, err := os.ReadDir(vulnsDir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", vulnsDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(vulnsDir, entry.Name())
		if filepath.Ext(filePath) == ".json" {
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", filePath, err)
			}
			err = LoadVulnerabilities(storage, content)
			if err != nil {
				t.Fatalf("Failed to load vulnerabilities from file %s: %v", filePath, err)
			}
		}
	}
	// Test ingestion of vulnerabilities

	err = Vulnerabilities(storage, nil)

	if err != nil {
		t.Fatal(err)
	}

	keys, err = storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	if len(keys) != numberOfNodes+3 {
		t.Fatalf("Expected number of nodes to be %d, got %d", numberOfNodes+3, len(keys))
	}
}
