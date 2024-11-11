package ingest

import (
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"
)

func TestGetPURLEcosystem(t *testing.T) {
	tests := []struct {
		name     string
		purl     packageurl.PackageURL
		expected Ecosystem
	}{
		{
			name:     "unknown type",
			purl:     packageurl.PackageURL{Type: "unknown", Namespace: "test"},
			expected: Ecosystem("unknown:test"),
		},
		{
			name:     "wildcard match",
			purl:     packageurl.PackageURL{Type: "npm", Namespace: "anything"},
			expected: EcosystemNPM,
		},
		{
			name:     "specific namespace match",
			purl:     packageurl.PackageURL{Type: "apk", Namespace: "alpine"},
			expected: EcosystemAlpine,
		},
		{
			name:     "namespace mismatch",
			purl:     packageurl.PackageURL{Type: "apk", Namespace: "wrong"},
			expected: Ecosystem("apk:wrong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPURLEcosystem(tt.purl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPURLToPackage(t *testing.T) {
	tests := []struct {
		name        string
		purl        string
		expected    PackageInfo
		expectError bool
	}{
		{
			name: "simple package without namespace",
			purl: "pkg:npm/express@4.17.1",
			expected: PackageInfo{
				Name:      "express",
				Version:   "4.17.1",
				Ecosystem: "npm",
			},
		},
		{
			name: "package with namespace",
			purl: "pkg:npm/%40babel/core@7.0.0",
			expected: PackageInfo{
				Name:      "@babel/core",
				Version:   "7.0.0",
				Ecosystem: "npm",
			},
		},
		{
			name: "maven package",
			purl: "pkg:maven/org.apache.commons/commons-lang3@3.12.0",
			expected: PackageInfo{
				Name:      "org.apache.commons:commons-lang3",
				Version:   "3.12.0",
				Ecosystem: "Maven",
			},
		},
		{
			name: "debian package",
			purl: "pkg:deb/debian/curl@7.74.0-1.3",
			expected: PackageInfo{
				Name:      "curl",
				Version:   "7.74.0-1.3",
				Ecosystem: "Debian",
			},
		},
		{
			name: "alpine package",
			purl: "pkg:apk/alpine/curl@7.74.0-r1",
			expected: PackageInfo{
				Name:      "curl",
				Version:   "7.74.0-r1",
				Ecosystem: "Alpine",
			},
		},
		{
			name:        "invalid purl",
			purl:        "invalid:purl",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PURLToPackage(tt.purl)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
} 