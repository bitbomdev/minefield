package ingest

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
)

func TestIngestSBOM(t *testing.T) {
	createTestFiles(t)
	tests := []struct {
		name     string
		sbomPath string
		want     map[uint32]*graph.Node
		wantErr  bool
	}{
		{
			name:     "default",
			sbomPath: "../../../test",
			want: map[uint32]*graph.Node{
				1: {
					ID:   1,
					Type: "library",
					Name: "pkg:generic/dep1",
				},
				2: {
					ID:   2,
					Type: "library",
					Name: "pkg:generic/dep1/subcomponent",
				},
				3: {
					ID:   3,
					Type: "library",
					Name: "pkg:generic/dep2",
				},
				4: {
					ID:   4,
					Type: "library",
					Name: "pkg:generic/lib-A",
				},
				5: {
					ID:   5,
					Type: "library",
					Name: "pkg:generic/lib-B",
				},
			},
		},
		{
			name:     "non-existent file",
			sbomPath: "non_existent_file.json",
			wantErr:  true,
		},
		{
			name:     "empty directory",
			sbomPath: "../../../empty_dir",
			want:     map[uint32]*graph.Node{},
			wantErr:  false,
		},
		{
			name:     "invalid SBOM file",
			sbomPath: "../../../invalid_sbom.json",
			wantErr:  true,
		},
		{
			name:     "SBOM with no components",
			sbomPath: "../../../no_components_sbom.json",
			want:     map[uint32]*graph.Node{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := graph.NewMockStorage()
			if _, err := SBOM(test.sbomPath, storage); test.wantErr != (err != nil) {
				t.Errorf("Sbom() error = %v, wantErr = %v", err, test.wantErr)
			}

			keys, err := storage.GetAllKeys()
			if err != nil {
				t.Fatalf("Failed to get all keys, %v", err)
			}

			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})

			for _, key := range keys {
				node, err := storage.GetNode(key)
				if err != nil {
					t.Fatalf("Failed to get node, %v", err)
				}
				fmt.Println(node.Name)
				if !nodeEquals(node, test.want[key]) {
					t.Fatalf("expected node %v, got %v", test.want[key], node)
				}
			}
		})
	}
	if err := os.RemoveAll("../../../empty_dir"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove("../../../invalid_sbom.json"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove("../../../no_components_sbom.json"); err != nil {
		t.Fatal(err)
	}
}

func nodeEquals(n, n2 *graph.Node) bool {
	if ((n == nil || n2 == nil) && n != n2) ||
		(n != nil && (n.ID != n2.ID || n.Type != n2.Type || n.Name != n2.Name)) {
		return false
	}
	return true
}

func createTestFiles(t *testing.T) {
	t.Helper()

	// Create an empty directory
	err := os.MkdirAll("../../../empty_dir", 0o755)
	if err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Create an invalid SBOM file
	invalidSBOM := []byte(`{"invalid": "json"}`)
	err = os.WriteFile("../../../invalid_sbom.json", invalidSBOM, 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid SBOM file: %v", err)
	}

	// Create a SBOM file with no components
	noComponentsSBOM := []byte(`{
		"bomFormat": "CycloneDX",
		"specVersion": "1.5",
		"version": 1,
		"metadata": {},
		"components": []
	}`)
	err = os.WriteFile("../../../no_components_sbom.json", noComponentsSBOM, 0o644)
	if err != nil {
		t.Fatalf("Failed to create no components SBOM file: %v", err)
	}
}
