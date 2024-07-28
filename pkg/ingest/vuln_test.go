package ingest

import (
	"testing"

	"github.com/bit-bom/bitbom/pkg"
	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"
)

func TestVulnerabilities(t *testing.T) {
	storage := pkg.NewMockStorage()
	// Add mock nodes to storage
	_, err := pkg.AddNode(storage, "PACKAGE", "metadata1", "pkg:golang/github.com/golang/go")
	assert.NoError(t, err)
	_, err = pkg.AddNode(storage, "PACKAGE", "metadata2", "pkg:cargo/github/rust-lang/rust")
	assert.NoError(t, err)

	err = Vulnerabilities(storage)
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
	}{
		{"pkg:golang/github.com/pkg/errors@v0.9.1", EcosystemGo},
		{"pkg:npm/lodash@4.17.20", EcosystemNPM},
		{"pkg:deb/debian/curl@7.68.0-1ubuntu2.6", EcosystemDebian},
		{"pkg:hex/phoenix@1.5.7", EcosystemHex},
	}

	for _, test := range tests {
		purl, err := packageurl.FromString(test.purl)
		assert.NoError(t, err)
		ecosystem := getPURLEcosystem(purl)
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
