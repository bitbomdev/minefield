package graph

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	storage := NewMockStorage()
	node, err := AddNode(storage, "type1", "metadata1", "name1")

	assert.NoError(t, err)
	pulledNode, err := storage.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, node, pulledNode, "Expected 1 node")
}

func TestSetDependency(t *testing.T) {
	storage := NewMockStorage()
	node1, err := AddNode(storage, "type1", "metadata1", "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", "name2")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)

	assert.NoError(t, err)
	assert.Contains(t, node1.Children.ToArray(), node2.ID, "Expected node1 to have node2 as child dependency")
	assert.Contains(t, node2.Parents.ToArray(), node1.ID, "Expected node2 to have node1 as parent dependency")
}

func TestSetDependent(t *testing.T) {
	storage := NewMockStorage()
	node1, err := AddNode(storage, "type1", "metadata1", "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", "name2")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)

	assert.NoError(t, err)
	assert.Contains(t, node2.Parents.ToArray(), node1.ID, "Expected node2 to have node1 as parent dependency")
}

func TestQueryDependentsAndDependenciesNoCache(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(storage Storage) (*Node, error)
		wantDependents   []uint32
		wantDependencies []uint32
		wantErr          bool
	}{
		{
			name: "empty graph",
			setup: func(storage Storage) (*Node, error) {
				return AddNode(storage, "type", "metadata", "empty")
			},
			wantDependents:   []uint32{1},
			wantDependencies: []uint32{1},
		},
		{
			name: "single parent-child relationship",
			setup: func(storage Storage) (*Node, error) {
				child, _ := AddNode(storage, "type", "metadata", "child")
				parent, _ := AddNode(storage, "type", "metadata", "parent")
				parent.SetDependency(storage, child)
				return parent, nil
			},
			wantDependents:   []uint32{2},
			wantDependencies: []uint32{1, 2},
		},
		{
			name: "complex graph",
			setup: func(storage Storage) (*Node, error) {
				n1, _ := AddNode(storage, "type", "metadata", "n1")
				n2, _ := AddNode(storage, "type", "metadata", "n2")
				n3, _ := AddNode(storage, "type", "metadata", "n3")
				n4, _ := AddNode(storage, "type", "metadata", "n4")
				n1.SetDependency(storage, n2)
				n2.SetDependency(storage, n3)
				n2.SetDependency(storage, n4)
				n3.SetDependency(storage, n4)
				return n2, nil
			},
			wantDependents:   []uint32{1, 2},
			wantDependencies: []uint32{2, 3, 4},
		},
		{
			name: "cyclic graph",
			setup: func(storage Storage) (*Node, error) {
				n1, _ := AddNode(storage, "type", "metadata", "n1")
				n2, _ := AddNode(storage, "type", "metadata", "n2")
				n3, _ := AddNode(storage, "type", "metadata", "n3")
				n1.SetDependency(storage, n2)
				n2.SetDependency(storage, n3)
				n3.SetDependency(storage, n1)
				return n1, nil
			},
			wantDependents:   []uint32{1, 2, 3},
			wantDependencies: []uint32{1, 2, 3},
		},
		{
			name: "deep dependency chain",
			setup: func(storage Storage) (*Node, error) {
				n1, _ := AddNode(storage, "type", "metadata", "n1")
				n2, _ := AddNode(storage, "type", "metadata", "n2")
				n3, _ := AddNode(storage, "type", "metadata", "n3")
				n4, _ := AddNode(storage, "type", "metadata", "n4")
				n5, _ := AddNode(storage, "type", "metadata", "n5")
				n1.SetDependency(storage, n2)
				n2.SetDependency(storage, n3)
				n3.SetDependency(storage, n4)
				n4.SetDependency(storage, n5)
				return n3, nil
			},
			wantDependents:   []uint32{1, 2, 3},
			wantDependencies: []uint32{3, 4, 5},
		},
		{
			name: "diamond dependency",
			setup: func(storage Storage) (*Node, error) {
				n1, _ := AddNode(storage, "type", "metadata", "n1")
				n2, _ := AddNode(storage, "type", "metadata", "n2")
				n3, _ := AddNode(storage, "type", "metadata", "n3")
				n4, _ := AddNode(storage, "type", "metadata", "n4")
				n1.SetDependency(storage, n2)
				n1.SetDependency(storage, n3)
				n2.SetDependency(storage, n4)
				n3.SetDependency(storage, n4)
				return n1, nil
			},
			wantDependents:   []uint32{1},
			wantDependencies: []uint32{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage()
			node, err := tt.setup(storage)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			gotDependents, err := node.QueryDependentsNoCache(storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryDependentsNoCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(gotDependents.ToArray(), tt.wantDependents) {
				t.Errorf("QueryDependentsNoCache() = %v, want %v", gotDependents.ToArray(), tt.wantDependents)
			}

			gotDependencies, err := node.QueryDependenciesNoCache(storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryDependenciesNoCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(gotDependencies.ToArray(), tt.wantDependencies) {
				t.Errorf("QueryDependenciesNoCache() = %v, want %v", gotDependencies.ToArray(), tt.wantDependencies)
			}
		})
	}
}

func TestNodeJSONMarshalUnmarshal(t *testing.T) {
	// Create a test Node
	node := &Node{
		ID:       1,
		Type:     "testType",
		Name:     "testName",
		Metadata: "testMetadata",
		Children: roaring.New(),
		Parents:  roaring.New(),
	}
	node.Children.AddMany([]uint32{5, 6, 7})
	node.Parents.AddMany([]uint32{2, 3, 4})

	// Test Node marshaling and unmarshaling
	nodeJSON, err := json.Marshal(node)
	assert.NoError(t, err, "Failed to marshal Node")

	var unmarshaledNode Node
	err = json.Unmarshal(nodeJSON, &unmarshaledNode)
	assert.NoError(t, err, "Failed to unmarshal Node")

	assert.Equal(t, node.ID, unmarshaledNode.ID)
	assert.Equal(t, node.Type, unmarshaledNode.Type)
	assert.Equal(t, node.Name, unmarshaledNode.Name)
	assert.Equal(t, node.Metadata, unmarshaledNode.Metadata)
	assert.True(t, node.Children.Equals(unmarshaledNode.Children))
	assert.True(t, node.Parents.Equals(unmarshaledNode.Parents))
}

func TestNodeCacheJSONMarshalUnmarshal(t *testing.T) {
	// Create a test NodeCache
	nodeCache := &NodeCache{
		ID:          1,
		AllParents:  roaring.New(),
		AllChildren: roaring.New(),
	}
	nodeCache.AllParents.AddMany([]uint32{5, 6, 7})
	nodeCache.AllChildren.AddMany([]uint32{2, 3, 4})

	// Test NodeCache marshaling and unmarshaling
	nodeCacheJSON, err := json.Marshal(nodeCache)
	assert.NoError(t, err, "Failed to marshal NodeCache")

	var unmarshaledNodeCache NodeCache
	err = json.Unmarshal(nodeCacheJSON, &unmarshaledNodeCache)
	assert.NoError(t, err, "Failed to unmarshal NodeCache")

	assert.Equal(t, nodeCache.ID, unmarshaledNodeCache.ID)
	assert.True(t, nodeCache.AllParents.Equals(unmarshaledNodeCache.AllParents))
	assert.True(t, nodeCache.AllChildren.Equals(unmarshaledNodeCache.AllChildren))
}
