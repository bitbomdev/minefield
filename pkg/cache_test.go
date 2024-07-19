package pkg

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findCycles(t *testing.T) {
	logger := log.Default()

	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", "1")
	assert.NoError(t, err)
	node2, err := AddNode(storage, "type2", "metadata2", "2")
	assert.NoError(t, err)
	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)

	got, err := findCycles[string](storage, "children", int(node2.Id), 2)
	if err != nil {
		logger.Fatalf("error finding cycles, storage %v, err %v", storage, err)
		return
	}

	if got.count != 2 {
		logger.Fatalf("findCycles want: %v unions, got: %v unions", 2, got.count)
	}
}

func Test_findCycles_With_Cycles(t *testing.T) {
	logger := log.Default()

	storage := NewMockStorage[string]()
	node1, err := AddNode(storage, "type1", "metadata1", "1")
	assert.NoError(t, err)
	node2, err := AddNode(storage, "type2", "metadata2", "2")
	assert.NoError(t, err)
	node3, err := AddNode(storage, "type3", "metadata3", "3")
	assert.NoError(t, err)

	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)
	err = node2.SetDependency(storage, node3)
	assert.NoError(t, err)
	err = node3.SetDependency(storage, node1)
	assert.NoError(t, err)

	got, err := findCycles[string](storage, "children", int(node2.Id), 3)
	if err != nil {
		logger.Fatalf("error finding cycles, storage %v, err %v", storage, err)
		return
	}

	if got.count != 1 {
		logger.Fatalf("findCycles want: %v unions, got: %v unions", 1, got.count)
	}
}
