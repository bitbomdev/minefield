package graph

import (
	"testing"

	"github.com/RoaringBitmap/roaring"
)

// TestParseAndExecute tests basic queries on simple mock data.
// We extend these tests with queries that include qualifiers (e.g., ?type=jar).
func TestParseAndExecute(t *testing.T) {
	storage := NewMockStorage()

	// Create some mock nodes.
	node1, err := AddNode(storage, "PACKAGE", nil, "pkg:generic/lib-A@1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	node2, err := AddNode(storage, "PACKAGE", nil, "pkg:generic/lib-B@1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	node3, err := AddNode(storage, "PACKAGE", nil, "pkg:generic/dep1@1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	node4, err := AddNode(storage, "PACKAGE", nil, "pkg:generic/dep2@1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Create a node with a query-string in its purl.
	nodeCamel, err := AddNode(storage, "PACKAGE", nil, "pkg:maven/org.apache.camel.quarkus/camel-quarkus-cassandraql-deployment@3.18.0-SNAPSHOT?type=jar")
	if err != nil {
		t.Fatal(err)
	}

	// Set up some dependencies.
	// Node1 -> Node3 -> Node4
	err = node1.SetDependency(storage, node3)
	if err != nil {
		t.Fatal(err)
	}
	err = node2.SetDependency(storage, node3)
	if err != nil {
		t.Fatal(err)
	}
	err = node3.SetDependency(storage, node4)
	if err != nil {
		t.Fatal(err)
	}

	// Make nodeCamel depend on node4 to verify a dependency query with a qualifier in the PURL.
	err = nodeCamel.SetDependency(storage, node4)
	if err != nil {
		t.Fatal(err)
	}

	// Cache the results for quicker lookups.
	if err := Cache(storage); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		script          string
		defaultNodeName string
		want            *roaring.Bitmap
		wantErr         bool
	}{
		{
			name:            "Simple dependents query",
			script:          "dependents PACKAGE pkg:generic/dep1@1.0.0",
			want:            roaring.BitmapOf(node1.ID, node2.ID, node3.ID),
			defaultNodeName: "",
		},
		{
			name:            "Simple dependencies query",
			script:          "dependencies PACKAGE pkg:generic/lib-A@1.0.0",
			want:            roaring.BitmapOf(node1.ID, node3.ID, node4.ID),
			defaultNodeName: "",
		},
		{
			name:            "Invalid token",
			script:          "invalid PACKAGE pkg:generic/lib-A@1.0.0",
			wantErr:         true,
			defaultNodeName: "",
		},
		{
			name:            "Combine dependents and dependencies with OR",
			script:          "dependents PACKAGE pkg:generic/dep1@1.0.0 or dependencies PACKAGE pkg:generic/dep1@1.0.0",
			want:            roaring.BitmapOf(node1.ID, node2.ID, node3.ID, node4.ID),
			defaultNodeName: "",
		},
		{
			name:            "Mismatched parentheses",
			script:          "dependents PACKAGE pkg:generic/lib-A@1.0.0 ]",
			wantErr:         true,
			defaultNodeName: "",
		},
		{
			name:            "Empty node name defaults to pkg:generic/lib-A@1.0.0",
			script:          "dependents PACKAGE or dependencies PACKAGE",
			want:            roaring.BitmapOf(node1.ID, node3.ID, node4.ID),
			defaultNodeName: "pkg:generic/lib-A@1.0.0",
		},

		// New tests exercising qualifiers in the PURL.
		{
			name:            "Dependencies query with ?type=jar qualifier",
			script:          "dependencies PACKAGE pkg:maven/org.apache.camel.quarkus/camel-quarkus-cassandraql-deployment@3.18.0-SNAPSHOT?type=jar",
			want:            roaring.BitmapOf(nodeCamel.ID, node4.ID),
			defaultNodeName: "",
		},
		{
			name:            "Dependents query with ?type=jar qualifier",
			script:          "dependents PACKAGE pkg:maven/org.apache.camel.quarkus/camel-quarkus-cassandraql-deployment@3.18.0-SNAPSHOT?type=jar",
			want:            roaring.BitmapOf(nodeCamel.ID),
			defaultNodeName: "",
		},
		{
			name:            "Invalid query with qualifier",
			script:          "invalid PACKAGE pkg:maven/org.apache.camel.quarkus/camel-quarkus-cassandraql@3.18.0-SNAPSHOT?type=jar",
			wantErr:         true,
			defaultNodeName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := storage.GetAllKeys()
			if err != nil {
				t.Fatal(err)
			}
			nodes, err := storage.GetNodes(keys)
			if err != nil {
				t.Fatal(err)
			}
			caches, err := storage.GetCaches(keys)
			if err != nil {
				t.Fatal(err)
			}

			result, err := ParseAndExecute(tt.script, storage, tt.defaultNodeName, nodes, caches, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAndExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.want) {
				t.Errorf("ParseAndExecute() got = %v, want %v", result, tt.want)
			}
		})
	}
}
