package pkg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring"
)

var ErrNodeAlreadyExists = errors.New("node with name already exists")

// Generic Node structure with metadata as generic type
type Node[T any] struct {
	Metadata   T               `json:"metadata"`
	Child      *roaring.Bitmap `json:"child"`
	Parent     *roaring.Bitmap `json:"parent"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	ChildData  []byte          `json:"childData"`
	ParentData []byte          `json:"parentData"`
	Id         uint32          `json:"Id"`
}

// MarshalJSON is a custom JSON marshalling tool.
// Roaring bitmaps can't be marshaled directly, so we need to call the roaring bitmaps function to convert the bitmaps to an []byte
// This takes the roaring bitmaps "Child" and "Parent" and converts them to byte slices called "ChildData" and "ParentData".
func (n *Node[T]) MarshalJSON() ([]byte, error) {
	childData, err := n.Child.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert child bitmap to bytes: %w", err)
	}
	parentData, err := n.Parent.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert parent bitmap to bytes: %w", err)
	}
	return json.Marshal(&struct {
		Metadata   T      `json:"metadata"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
		ID         uint32 `json:"Id"`
	}{
		ID:         n.Id,
		Type:       n.Type,
		Name:       n.Name,
		Metadata:   n.Metadata,
		ChildData:  childData,
		ParentData: parentData,
	})
}

// UnmarshalJSON is a custom JSON unmarshalling tool.
// We store the roaring bitmaps as a byte slice, so we need to unmarshal them, and then convert them from []byte to roaring.Bitmap.
// This takes the "ChildData" and "ParentData" fields and unmarshal them from bytes into roaring bitmaps.
func (n *Node[T]) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Metadata   T      `json:"metadata"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
		ID         uint32 `json:"Id"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal node data: %w", err)
	}
	n.Id = aux.ID
	n.Type = aux.Type
	n.Name = aux.Name
	n.Metadata = aux.Metadata
	n.Child = roaring.New()
	n.Parent = roaring.New()
	if _, err := n.Child.FromBuffer(aux.ChildData); err != nil {
		return fmt.Errorf("failed to convert child data from buffer: %w", err)
	}
	if _, err := n.Parent.FromBuffer(aux.ParentData); err != nil {
		return fmt.Errorf("failed to convert parent data from buffer: %w", err)
	}
	return nil
}

// AddNode becomes generic in terms of metadata
func AddNode[T any](storage Storage[T], _type string, metadata T, parent, child *roaring.Bitmap, name string) (*Node[T], error) {
	var ID uint32
	if id, err := storage.NameToID(name); err == nil {
		return storage.GetNode(id)
	} else {
		ID, err = storage.GenerateID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	n := &Node[T]{
		Id:       ID,
		Type:     _type,
		Name:     name,
		Metadata: metadata,
		Child:    child,
		Parent:   parent,
	}
	if err := storage.SaveNode(n); err != nil {
		return nil, fmt.Errorf("failed to save node: %w", err)
	}
	return n, nil
}

// SetDependency now uses generic types for metadata
func (n *Node[T]) SetDependency(storage Storage[T], neighbor *Node[T]) error {
	if n == nil {
		return fmt.Errorf("cannot add dependency to nil node")
	}
	if neighbor == nil {
		return fmt.Errorf("cannot add dependency to nil node")
	}
	if n.Id == neighbor.Id {
		return fmt.Errorf("cannot add self as dependency")
	}

	n.Child.Add(neighbor.Id)
	neighbor.Parent.Add(n.Id)

	if err := storage.SaveNode(n); err != nil {
		return fmt.Errorf("failed to save node: %w", err)
	}
	if err := storage.SaveNode(neighbor); err != nil {
		return fmt.Errorf("failed to save neighbor node: %w", err)
	}
	return nil
}

func (n *Node[T]) queryBitmap(storage Storage[T], direction string) (*roaring.Bitmap, error) {
	if n == nil {
		return nil, fmt.Errorf("cannot query bitmap of nil node")
	}

	result := roaring.New()
	visited := make(map[uint32]bool)
	queue := []*Node[T]{n}

	for len(queue) > 0 {
		curNode := queue[0]
		queue = queue[1:]

		if visited[curNode.Id] {
			continue
		}
		visited[curNode.Id] = true

		var bitmap *roaring.Bitmap
		switch direction {
		case "child":
			bitmap = curNode.Child
		case "parent":
			bitmap = curNode.Parent
		default:
			return nil, fmt.Errorf("invalid direction during query: %s", direction)
		}

		result.Or(bitmap)
		for _, nID := range bitmap.Clone().ToArray() {
			node, err := storage.GetNode(nID)
			if err != nil {
				return nil, fmt.Errorf("failed to get node: %w", err)
			}
			queue = append(queue, node)
		}
	}

	return result, nil
}

func (n *Node[T]) QueryDependents(storage Storage[T]) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, "parent")
}

func (n *Node[T]) QueryDependencies(storage Storage[T]) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, "child")
}
