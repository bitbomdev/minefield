package pkg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring"
)

// Generic Node structure with metadata as generic type
type Node[T any] struct {
	Id         uint32         `json:"Id"`
	Type       string         `json:"type"`
	Metadata   T              `json:"metadata"`
	Child      roaring.Bitmap `json:"child"`
	Parent     roaring.Bitmap `json:"parent"`
	ChildData  []byte         `json:"childData"`
	ParentData []byte         `json:"parentData"`
}

// MarshalJSON is a custom JSON marshalling tool.
// Roaring bitmaps can't be marshaled directly, so we need to call the roaring bitmaps function to convert the bitmaps to an []byte
// This takes the roaring bitmaps "Child" and "Parent" and converts them to byte slices called "ChildData" and "ParentData".
func (n *Node[T]) MarshalJSON() ([]byte, error) {
	childData, err := n.Child.ToBytes()
	if err != nil {
		return nil, err
	}
	parentData, err := n.Parent.ToBytes()
	if err != nil {
		return nil, err
	}
	return json.Marshal(&struct {
		ID         uint32 `json:"Id"`
		Type       string `json:"type"`
		Metadata   T      `json:"metadata"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
	}{
		ID:         n.Id,
		Type:       n.Type,
		Metadata:   n.Metadata,
		ChildData:  childData,
		ParentData: parentData,
	})
}

// UnmarshalJSON is a custom JSON unmarshaling tool.
// We store the roaring bitmaps as a byte slice, so we need to unmarshal them, and then convert them from []byte to roaring.Bitmap.
// This takes the "ChildData" and "ParentData" fields and unmarshals them from bytes into roaring bitmaps.
func (n *Node[T]) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ID         uint32 `json:"Id"`
		Type       string `json:"type"`
		Metadata   T      `json:"metadata"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	n.Id = aux.ID
	n.Type = aux.Type
	n.Metadata = aux.Metadata
	if _, err := n.Child.FromBuffer(aux.ChildData); err != nil {
		return err
	}
	if _, err := n.Parent.FromBuffer(aux.ParentData); err != nil {
		return err
	}
	return nil
}

// AddNode becomes generic in terms of metadata
func AddNode[T any](storage Storage[T], _type string, metadata T, parent, child roaring.Bitmap) (*Node[T], error) {
	ID, err := storage.GenerateID()
	if err != nil {
		return nil, err
	}
	n := &Node[T]{
		Id:       ID,
		Type:     _type,
		Metadata: metadata,
		Child:    child,
		Parent:   parent,
	}
	if err := storage.SaveNode(n); err != nil {
		return nil, err
	}
	return n, nil
}

// SetDependency now uses generic types for metadata
func (n *Node[T]) SetDependency(storage Storage[T], neighbor *Node[T]) error {
	if n == nil {
		return errors.New("cannot add dependency to nil node")
	}
	if neighbor == nil {
		return errors.New("cannot add dependency to nil node")
	}
	if n.Id == neighbor.Id {
		return errors.New("cannot add self as dependency")
	}

	n.Child.Add(neighbor.Id)
	neighbor.Parent.Add(n.Id)

	if err := storage.SaveNode(n); err != nil {
		return err
	}
	if err := storage.SaveNode(neighbor); err != nil {
		return err
	}
	return nil
}

func (n *Node[T]) queryBitmap(storage Storage[T], direction string) (*roaring.Bitmap, error) {
	if n == nil {
		return nil, errors.New("cannot query bitmap of nil node")
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
			bitmap = &curNode.Child
		case "parent":
			bitmap = &curNode.Parent
		default:
			return nil, fmt.Errorf("invalID direction during query: %s", direction)
		}

		result.Or(bitmap)
		for _, nID := range bitmap.Clone().ToArray() {
			node, err := storage.GetNode(nID)
			if err != nil {
				return nil, err
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
