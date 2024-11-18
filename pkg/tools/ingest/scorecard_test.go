package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitbomdev/minefield/pkg/graph"
)

func TestScorecards(t *testing.T) {
	storage := graph.NewMockStorage()

	scorecardsDir := "../../../testdata/scorecards"
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
			t.Fatalf("Failed to load SBOM from file %s: %v", file.Name(), err)
		}
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	numberOfNodes := len(keys)

	// Read scorecard files
	scorecardFiles, err := os.ReadDir(scorecardsDir)
	if err != nil {
		t.Fatalf("Failed to read scorecards directory: %v", err)
	}

	scorecardCount := 0
	for _, file := range scorecardFiles {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(scorecardsDir, file.Name()))
		if err != nil {
			t.Fatalf("Failed to read scorecard file %s: %v", file.Name(), err)
		}

		if err := Scorecards(storage, data); err != nil {
			t.Fatalf("Failed to load scorecard from file %s: %v", file.Name(), err)
		}
		scorecardCount++
	}

	if scorecardCount == 0 {
		t.Fatal("Expected scorecards to be ingested, got 0")
	}

	keys, err = storage.GetAllKeys()
	if err != nil {
		t.Fatalf("Failed to get all keys, %v", err)
	}

	if len(keys) != numberOfNodes+1 {
		t.Fatalf("Expected number of nodes to be %d, got %d", numberOfNodes+1, len(keys))
	}
}
