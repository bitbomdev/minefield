package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitbomdev/minefield/pkg/graph"
)

func TestIngestSBOM(t *testing.T) {
	storage := graph.NewMockStorage()

	sbomDir := "../../../testdata/sboms"

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
			t.Fatalf("Failed to process SBOM from file %s: %v", file.Name(), err)
		}
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys: %v", err)
	}

	// Verify we have the expected number of nodes
	if len(keys) != 1600 {
		t.Fatalf("Expected 1600 nodes to be created from SBOM ingestion, got %d", len(keys))
	}

}
