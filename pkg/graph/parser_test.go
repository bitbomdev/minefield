package graph

import (
	"testing"

	"github.com/RoaringBitmap/roaring"
)

func TestParseAndExecute(t *testing.T) {
	storage := NewMockStorage()

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
			want:            roaring.BitmapOf(1, 2, 3),
			defaultNodeName: "",
		},
		{
			name:            "Simple dependencies query",
			script:          "dependencies PACKAGE pkg:generic/lib-A@1.0.0",
			want:            roaring.BitmapOf(1, 3, 4),
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
			want:            roaring.BitmapOf(1, 2, 3, 4),
			defaultNodeName: "",
		},
		{
			name:            "Mismatched parentheses",
			script:          "dependents PACKAGE pkg:generic/lib-A@1.0.0 ]",
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
