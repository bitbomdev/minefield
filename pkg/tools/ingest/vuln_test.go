package ingest

import (
	"testing"

	"github.com/bitbomdev/minefield/pkg/graph"
)

func TestVulnerabilities(t *testing.T) {
	storage := graph.NewMockStorage()

	vulnsDir := "../../../testdata/osv-vulns"
	sbomDir := "../../../testdata/osv-sboms"

	result, err := LoadDataFromPath(storage, sbomDir)
	if err != nil {
		t.Fatalf("Failed to ingest SBOM: %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("Expected SBOM to be ingested, got %d", len(result))
	}

	for _, data := range result {
		if err := SBOM(storage, data.Data); err != nil {
			t.Fatalf("Failed to load SBOM from data: %v", err)
		}
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	numberOfNodes := len(keys)

	result, err = LoadDataFromPath(storage, vulnsDir)
	if err != nil {
		t.Fatalf("Failed to load vulnerabilities from directory %s: %v", vulnsDir, err)
	}
	if len(result) == 0 {
		t.Fatalf("Expected vulnerabilities to be ingested, got %d", len(result))
	}
	for _, data := range result {
		if err := Vulnerabilities(storage, data.Data); err != nil {
			t.Fatalf("Failed to load vulnerabilities from data: %v", err)
		}
	}

	keys, err = storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	if len(keys) != numberOfNodes+3 {
		t.Fatalf("Expected number of nodes to be %d, got %d", numberOfNodes+3, len(keys))
	}
}
