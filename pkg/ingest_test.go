package pkg

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/protobom/protobom/pkg/sbom"
	"sort"
	"testing"
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
			sbomPath: "../test-data/syft-cyclonedx-docker.io-library-busybox.latest.json",
			want: func() map[uint32]*Node[any] {
				bitmap1Child := roaring.New()
				bitmap1Child.Add(2)

				bitmap2Parent := roaring.New()
				bitmap2Parent.Add(1)

				return map[uint32]*Node[any]{
					1: &Node[any]{
						Id:       1,
						Type:     "empty_type",
						Metadata: "empty_metadata",
						Child:    *bitmap1Child,
					},
					2: &Node[any]{
						Id:       2,
						Type:     "empty_type",
						Metadata: "empty_metadata",
						Parent:   *bitmap2Parent,
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

func TestAddDependency(t *testing.T) {
	tests := []struct {
		name        string
		fromNodeIDs map[uint32][]uint32 // Map of fromNodeID to a slice of toNodeIDs
		wantErr     bool
	}{
		{
			name: "default",
			fromNodeIDs: map[uint32][]uint32{
				1: {2},
			},
			wantErr: false,
		},
		{
			name: "multiple dependencies",
			fromNodeIDs: map[uint32][]uint32{
				1: {2, 3},
				4: {5},
			},
			wantErr: false,
		},
		{
			name: "single node with no dependencies",
			fromNodeIDs: map[uint32][]uint32{
				1: {},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			fromNodeIDs: map[uint32][]uint32{
				1: {2},
				2: {1},
			},
			wantErr: false,
		},
		{
			name: "complex dependencies",
			fromNodeIDs: map[uint32][]uint32{
				1: {2, 3},
				2: {4},
				3: {4, 5},
				5: {6},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the storage
			storage := NewMockStorage[any]()

			// Create the document and protoIDToNodeID map
			document := &sbom.Document{}
			nodeList := &sbom.NodeList{}
			protoIDToNodeID := make(map[string]uint32)

			// Add nodes to the document and storage
			for fromNodeID, toNodeIDs := range test.fromNodeIDs {
				fromNodeIDStr := fmt.Sprintf("%d", fromNodeID)

				nodeList.Nodes = append(nodeList.Nodes, &sbom.Node{Id: fromNodeIDStr})
				protoIDToNodeID[fromNodeIDStr] = fromNodeID

				err := storage.SaveNode(&Node[any]{Id: fromNodeID})
				if err != nil {
					t.Fatalf("error saving node %v", err)
				}

				for _, toNodeID := range toNodeIDs {
					toNodeIDStr := fmt.Sprintf("%d", toNodeID)

					nodeList.Nodes = append(nodeList.Nodes, &sbom.Node{Id: toNodeIDStr})
					protoIDToNodeID[toNodeIDStr] = toNodeID

					err := storage.SaveNode(&Node[any]{Id: toNodeID})
					if err != nil {
						t.Fatalf("error saving node %v", err)
					}
				}
			}

			// Add edges to the document
			for fromNodeID, toNodeIDs := range test.fromNodeIDs {
				fromNodeIDStr := fmt.Sprintf("%d", fromNodeID)
				var toNodeIDStrs []string

				for _, toNodeID := range toNodeIDs {
					toNodeIDStrs = append(toNodeIDStrs, fmt.Sprintf("%d", toNodeID))
				}

				nodeList.Edges = append(nodeList.Edges, &sbom.Edge{From: fromNodeIDStr, To: toNodeIDStrs})
			}
			document.NodeList = nodeList

			// Call addDependency
			err := addDependency(document, storage, protoIDToNodeID)
			if test.wantErr != (err != nil) {
				t.Errorf("addDependency() error = %v, wantErr = %v", err, test.wantErr)
			}

			// Verify the dependencies
			for fromNodeID, toNodeIDs := range test.fromNodeIDs {
				fromNode, err := storage.GetNode(fromNodeID)
				if err != nil {
					t.Fatalf("Failed to get node %d, %v", fromNodeID, err)
				}
				for _, toNodeID := range toNodeIDs {
					toNode, err := storage.GetNode(toNodeID)
					if err != nil {
						t.Fatalf("Failed to get node %d, %v", toNodeID, err)
					}

					if !fromNode.Child.Contains(toNode.Id) {
						t.Fatalf("expected node %d to have dependency on node %d", fromNodeID, toNodeID)
					}
				}
			}
		})
	}
}
