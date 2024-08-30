package ingest

import (
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"
)

func TestVulnerabilities(t *testing.T) {
	storage := graph.NewMockStorage()
	// Add mock nodes to storages
	_, err := graph.AddNode(storage, "library", "metadata1", "pkg:golang/stdlib")
	assert.NoError(t, err)
	_, err = graph.AddNode(storage, "library", "metadata2", "pkg:golang/github/docker/docker@19.0.0")
	assert.NoError(t, err)

	err = Vulnerabilities(storage, nil)
	assert.NoError(t, err)

	// Check if vulnerabilities were added
	keys, err := storage.GetAllKeys()
	assert.NoError(t, err)
	assert.Greater(t, len(keys), 2) // Should have more than the initial 2 nodes
}

func TestGetPURLEcosystem(t *testing.T) {
	tests := []struct {
		purl     string
		expected Ecosystem
		wantErr  bool
	}{
		{"pkg:golang/github.com/pkg/errors@v0.9.1", EcosystemGo, false},
		{"pkg:npm/lodash@4.17.20", EcosystemNPM, false},
		{"pkg:deb/debian/curl@7.68.0-1ubuntu2.6", EcosystemDebian, false},
		{"pkg:hex/phoenix@1.5.7", EcosystemHex, false},
	}

	for _, test := range tests {
		purl, err := packageurl.FromString(test.purl)
		assert.NoError(t, err)
		ecosystem, err := getPURLEcosystem(purl)
		if (err != nil) != test.wantErr {
			t.Fatalf("getPURLEcosystem(%s) error = %v, wantErr %v", test.purl, err, test.wantErr)
		}
		assert.Equal(t, test.expected, ecosystem)
	}
}

func TestPURLToPackageQuery(t *testing.T) {
	tests := []struct {
		purl     string
		expected Query
	}{
		{
			"pkg:golang/github.com/pkg/errors@v0.9.1",
			Query{
				Version: "v0.9.1",
				Package: Package{
					Name:      "github.com/pkg/errors",
					Ecosystem: "Go",
				},
			},
		},
		{
			"pkg:npm/lodash@4.17.20",
			Query{
				Version: "4.17.20",
				Package: Package{
					Name:      "lodash",
					Ecosystem: "npm",
				},
			},
		},
	}

	for _, test := range tests {
		query, err := PURLToPackageQuery(test.purl)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, query)
	}
}
