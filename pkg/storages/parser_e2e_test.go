package storages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/stretchr/testify/assert"
)

func TestParseAndExecute_E2E(t *testing.T) {
	if _, ok := os.LookupEnv("e2e"); !ok {
		t.Skip("E2E tests are not enabled")
	}
	redisStorage := setupTestRedis()

	sbomPath := filepath.Join("..", "..", "test", "sboms")

	// Ingest data from the folder
	progress := func(count int, path string) {
		fmt.Printf("Ingested %d items from the folder %s\n", count, path)
	}
	count, err := ingest.SBOM(sbomPath, redisStorage, progress)
	assert.NoError(t, err)
	assert.Greater(t, count, 0)

	// Cache data
	err = graph.Cache(redisStorage, nil, nil)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		script          string
		defaultNodeName string
		want            *roaring.Bitmap
		wantErr         bool
	}{
		{
			name:            "Simple dependents query",
			script:          "dependents library pkg:generic/checkout",
			want:            roaring.BitmapOf(193, 196, 208, 309, 1214, 1265, 1276, 1308, 130),
			defaultNodeName: "",
		},
		{
			name:            "Simple dependencies query",
			script:          "dependencies library pkg:generic/checkout",
			want:            roaring.BitmapOf(193),
			defaultNodeName: "",
		},
		{
			name:            "Dependents query with xor",
			script:          "dependents library pkg:generic/checkout xor dependents library pkg:generic/setup-go",
			want:            roaring.BitmapOf(195, 196, 309, 130),
			defaultNodeName: "",
		},
		{
			name:            "Dependents query with and",
			script:          "dependents library pkg:generic/checkout and dependents library pkg:generic/setup-go",
			want:            roaring.BitmapOf(193, 208, 1214, 1265, 1276, 130),
			defaultNodeName: "",
		},
		{
			name:            "Empty script",
			script:          "",
			want:            roaring.New(),
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Invalid script",
			script:          "invalid script",
			want:            nil,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Complex nested expressions and not found",
			script:          "(dependents library pkg:generic/checkout and dependents library pkg:generic/setup-go) or dependents library pkg:generic/setup-go",
			want:            roaring.BitmapOf(130, 193, 208, 400, 401, 1214, 1265),
			defaultNodeName: "",
			wantErr:         false,
		},
		{
			name:            "Unknown query type",
			script:          "unknown library pkg:generic/checkout",
			want:            nil,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Missing node name",
			script:          "dependents library",
			want:            roaring.New(),
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Large input",
			script:          "dependents library pkg:generic/large",
			want:            roaring.BitmapOf(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Dependencies with OR operation",
			script:          "dependencies library pkg:generic/checkout or dependencies library pkg:generic/setup-go",
			want:            roaring.BitmapOf(193, 194),
			defaultNodeName: "",
		},
		{
			name:            "Nested expressions with parentheses",
			script:          "(dependencies library pkg:generic/checkout and dependents library pkg:generic/setup-go) or (dependencies library pkg:generic/setup-g and dependents library pkg:generic/checkout)",
			want:            roaring.BitmapOf(193, 208, 1214, 1265),
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Query with default node name",
			script:          "dependencies library",
			want:            roaring.BitmapOf(193),
			defaultNodeName: "pkg:generic/checkout",
			wantErr:         false,
		},
		{
			name:            "Complex query with multiple operations",
			script:          "((dependencies library pkg:generic/checkout or dependents library pkg:generic/setup-go) and dependencies library pkg:generic/setup-go) xor dependents library pkg:generic/checkout",
			want:            roaring.BitmapOf(196, 309, 1308, 130, 195, 194, 193, 208, 1214),
			defaultNodeName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := redisStorage.GetAllKeys()
			assert.NoError(t, err)

			nodes, err := redisStorage.GetNodes(keys)
			assert.NoError(t, err)

			caches, err := redisStorage.GetCaches(keys)
			assert.NoError(t, err)

			result, err := graph.ParseAndExecute(tt.script, redisStorage, tt.defaultNodeName, nodes, caches, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAndExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.GetCardinality() != tt.want.GetCardinality() {
				t.Errorf("ParseAndExecute() got = %v, want %v", result.ToArray(), tt.want.ToArray())
				t.Errorf("ParseAndExecute() got cardinality = %v, want cardinality %v", result.GetCardinality(), tt.want.GetCardinality())
			}
		})
	}
}
