package pkg

import (
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring"
)

// Generic Node structure with metadata as generic type
type Node[T any] struct {
	id       uint32
	_type    string
	metadata T
	child    roaring.Bitmap // what the node depends on
	parent   roaring.Bitmap // what depends on the node
}

func (n *Node[T]) GetID() uint32 { return n.id }

func (n *Node[T]) GetChild() roaring.Bitmap { return n.child }

func (n *Node[T]) GetParent() roaring.Bitmap { return n.parent }

// AddNode becomes generic in terms of metadata
func AddNode[T any](storage Storage[T], _type string, metadata T, parent, child roaring.Bitmap) (*Node[T], error) {
	id, err := storage.GenerateID()
	if err != nil {
		return nil, err
	}
	n := &Node[T]{
		id:       id,
		_type:    _type,
		metadata: metadata,
		child:    child,
		parent:   parent,
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
	if n.id == neighbor.id {
		return errors.New("cannot add self as dependency")
	}

	n.child.Add(neighbor.id)
	neighbor.parent.Add(n.id)

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

		if visited[curNode.id] {
			continue
		}
		visited[curNode.id] = true

		var bitmap *roaring.Bitmap
		switch direction {
		case "child":
			bitmap = &curNode.child
		case "parent":
			bitmap = &curNode.parent
		default:
			return nil, fmt.Errorf("invalid direction during query: %s", direction)
		}

		result.Or(bitmap)
		for _, nid := range bitmap.Clone().ToArray() {
			node, err := storage.GetNode(nid)
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
