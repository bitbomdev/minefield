package storages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/stretchr/testify/assert"
)

func TestParseAndExecute_E2E(t *testing.T) {
	//if _, ok := os.LookupEnv("e2e"); !ok {
	//	t.Skip("E2E tests are not enabled")
	//}
	redisStorage := setupTestRedis()

	sbomPath := filepath.Join("..", "..", "testdata", "sboms")
	vulnsPath := filepath.Join("..", "..", "testdata", "osv-vulns")

	// Ingest data from the folder
	progress := func(count int, path string) {
		fmt.Printf("Ingested %d items from the folder %s\n", count, path)
	}
	count, err := ingest.SBOM(sbomPath, redisStorage, progress)
	assert.NoError(t, err)
	assert.Greater(t, count, 0)
	err = filepath.Walk(vulnsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			err = ingest.LoadVulnerabilities(redisStorage, fileContent)
			if err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)

	err = ingest.Vulnerabilities(redisStorage, progress)
	assert.NoError(t, err)

	// Cache data
	err = graph.Cache(redisStorage)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		script          string
		defaultNodeName string
		want            uint64
		wantErr         bool
	}{
		{
			name:            "Simple dependents query",
			script:          "dependents library pkg:github/actions/checkout@v3",
			want:            6,
			defaultNodeName: "",
		},
		{
			name:            "Simple dependencies query",
			script:          "dependencies library pkg:github/actions/checkout@v3",
			want:            1,
			defaultNodeName: "",
		},
		{
			name:            "Dependents query with xor",
			script:          "dependents library pkg:github/actions/checkout@v3 xor dependents library pkg:golang/gopkg.in/yaml.v3@v3.0.1",
			want:            14,
			defaultNodeName: "",
		},
		{
			name:            "Dependents query with and",
			script:          "dependents library pkg:github/actions/checkout@v3 and dependents library pkg:golang/gopkg.in/yaml.v3@v3.0.1",
			want:            0,
			defaultNodeName: "",
		},
		{
			name:            "Empty script",
			script:          "",
			want:            0,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Invalid script",
			script:          "invalid script",
			want:            0,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Complex nested expressions",
			script:          "(dependents library pkg:github/actions/checkout@v3 and dependents library pkg:golang/gopkg.in/yaml.v3@v3.0.1) or dependents library pkg:golang/gopkg.in/yaml.v3@v3.0.1",
			want:            8,
			defaultNodeName: "",
			wantErr:         false,
		},
		{
			name:            "Unknown query type",
			script:          "unknown library pkg:github/actions/checkout@v3",
			want:            0,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Missing node name",
			script:          "dependents library",
			want:            0,
			defaultNodeName: "",
			wantErr:         true,
		},
		{
			name:            "Dependencies with OR operation",
			script:          "dependencies library pkg:github/actions/checkout@v3 or dependencies library pkg:golang/gopkg.in/yaml.v3@v3.0.1",
			want:            2,
			defaultNodeName: "",
		},
		{
			name:            "Query with default node name",
			script:          "dependencies library",
			want:            1,
			defaultNodeName: "pkg:github/actions/checkout@v3",
			wantErr:         false,
		},
		{
			name:            "Complex query with multiple operations",
			script:          "((dependencies library pkg:github/actions/checkout@v3 or dependents library pkg:golang/gopkg.in/yaml.v3@v3.0.1) and dependencies library pkg:golang/gopkg.in/yaml.v3@v3.0.1) xor dependents library pkg:github/actions/checkout@v3",
			want:            6,
			defaultNodeName: "",
		},
		{
			name:            "Query with vulnerability",
			script:          "dependencies vuln pkg:github.com/google/agi@",
			want:            1,
			defaultNodeName: "",
			wantErr:         false,
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
			if !tt.wantErr && result.GetCardinality() != tt.want {
				t.Errorf("ParseAndExecute() got cardinality = %v, want cardinality %v", result.GetCardinality(), tt.want)
			}
		})
	}
}
