package pkg

import (
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")

	assert.NoError(t, err)
	pulledNode, err := storage.GetNode(node.Id)
	assert.NoError(t, err)
	assert.Equal(t, node, pulledNode, "Expected 1 node")
}

func TestSetDependency(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", parent, child, "name2")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)

	assert.NoError(t, err)
	assert.Contains(t, node1.Child.ToArray(), node2.Id, "Expected node1 to have node2 as child dependency")
	assert.Contains(t, node2.Parent.ToArray(), node1.Id, "Expected node2 to have node1 as parent dependency")
}

func TestSetDependent(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", parent, child, "name2")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)

	assert.NoError(t, err)
	assert.Contains(t, node2.Parent.ToArray(), node1.Id, "Expected node2 to have node1 as parent dependency")
}

func TestQueryDependents(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", parent, child, "name2")
	assert.NoError(t, err, "Expected no error")

	err = node2.SetDependency(storage, node1)
	assert.NoError(t, err)
	dependents, err := node1.QueryDependents(storage)
	assert.NoError(t, err, "Expected no error")
	assert.Contains(t, dependents.ToArray(), node2.Id, "Expected node1 to have node2 as dependent")
}

func TestQueryDependencies(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", parent, child, "name2")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)
	dependencies, err := node1.QueryDependencies(storage)
	assert.NoError(t, err, "Expected no error")
	assert.Contains(t, dependencies.ToArray(), node2.Id, "Expected node1 to have node2 as dependency")
}

func TestCircularDependency(t *testing.T) {
	parent := roaring.New()
	child := roaring.New()
	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", parent, child, "name1")
	assert.NoError(t, err, "Expected no error")
	node2, err := AddNode(storage, "type2", "metadata2", parent, child, "name2")
	assert.NoError(t, err, "Expected no error")
	node3, err := AddNode(storage, "type3", "metadata3", parent, child, "name3")
	assert.NoError(t, err, "Expected no error")

	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)
	err = node2.SetDependency(storage, node3)
	assert.NoError(t, err)
	err = node3.SetDependency(storage, node1)
	assert.NoError(t, err)

	// Test QueryDependents for circular dependency
	dependents, err := node1.QueryDependents(storage)
	assert.NoError(t, err, "Expected no error")
	assert.Contains(t, dependents.ToArray(), node2.Id, "Expected node1 to have node2 as dependent")
	assert.Contains(t, dependents.ToArray(), node3.Id, "Expected node1 to have node3 as dependent")

	// Test QueryDependencies for circular dependency
	dependencies, err := node1.QueryDependencies(storage)
	assert.NoError(t, err, "Expected no error")
	t.Logf("Dependencies of node1: %v", dependencies.ToArray())
	assert.Contains(t, dependencies.ToArray(), node2.Id, "Expected node1 to have node2 as dependency")
	assert.Contains(t, dependencies.ToArray(), node3.Id, "Expected node1 to have node3 as dependency")
}
