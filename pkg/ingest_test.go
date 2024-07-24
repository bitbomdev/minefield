package pkg

import (
	"sort"
	"testing"

	"github.com/RoaringBitmap/roaring"
)

func TestIngestSBOM(t *testing.T) {
	tests := []struct {
		name     string
		sbomPath string
		want     map[uint32]*Node[any]
		wantErr  bool
	}{
		{
			name:     "default",
			sbomPath: "../test",
			want: func() map[uint32]*Node[any] {
				bitmap1Child := roaring.New()
				bitmap1Child.Add(2)

				bitmap2Parent := roaring.New()
				bitmap2Parent.Add(1)

				return map[uint32]*Node[any]{
					1: {
						Id:   1,
						Type: "PACKAGE",
						Name: "pkg:generic/dep1@1.0.0",
					},
					2: {
						Id:   2,
						Type: "PACKAGE",
						Name: "pkg:generic/dep2@1.0.0",
					},
					3: {
						Id:   3,
						Type: "PACKAGE",
						Name: "pkg:generic/lib-A@1.0.0",
					},
					4: {
						Id:   4,
						Type: "PACKAGE",
						Name: "pkg:generic/lib-B@1.0.0",
					},
				}
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMockStorage[any]()
			if err := IngestSBOM(test.sbomPath, storage); test.wantErr != (err != nil) {
				t.Errorf("IngestSBOM() error = %v, wantErr = %v", err, test.wantErr)
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

				if !node.nodeEquals(test.want[key]) {
					t.Fatalf("expected node %v, got %v", test.want[key], node)
				}
			}
		})
	}
}

func (n *Node[T]) nodeEquals(n2 *Node[T]) bool {
	if ((n == nil || n2 == nil) && n != n2) ||
		(n != nil && (n.Id != n2.Id || n.Type != n2.Type)) {
		return false
	}
	return true
}
