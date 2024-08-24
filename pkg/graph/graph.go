package graph

import (
	"errors"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/RoaringBitmap/roaring"
)

var (
	ErrNodeAlreadyExists = errors.New("node with name already exists")
	ErrSelfDependency    = errors.New("cannot add self as dependency")
)

type Direction string

const (
	ParentsDirection  Direction = "parents"
	ChildrenDirection Direction = "children"
)

// Generic Node structure with metadata as generic type
type Node struct {
	Metadata   any             `json:"metadata"`
	Children   *roaring.Bitmap `json:"child"`
	Parents    *roaring.Bitmap `json:"parent"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	ChildData  []byte          `json:"childData"`
	ParentData []byte          `json:"parentData"`
	ID         uint32          `json:"ID"`
}

type NodeCache struct {
	AllParents  *roaring.Bitmap
	AllChildren *roaring.Bitmap
	ID          uint32
}

func NewNodeCache(id uint32, allParents, allChildren *roaring.Bitmap) *NodeCache {
	return &NodeCache{
		ID:          id,
		AllParents:  allParents,
		AllChildren: allChildren,
	}
}

// MarshalJSON is a custom JSON marshalling method for NodeCache.
// It converts the roaring bitmaps to byte slices for JSON serialization.
func (nc *NodeCache) MarshalJSON() ([]byte, error) {
	allParentsData, err := nc.AllParents.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert AllParents bitmap to bytes: %w", err)
	}
	allChildrenData, err := nc.AllChildren.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert AllChildren bitmap to bytes: %w", err)
	}
	return json.Marshal(&struct {
		AllParentsData  []byte `json:"allParentsData"`
		AllChildrenData []byte `json:"allChildrenData"`
		NodeID          uint32 `json:"ID"`
	}{
		NodeID:          nc.ID,
		AllParentsData:  allParentsData,
		AllChildrenData: allChildrenData,
	})
}

// UnmarshalJSON is a custom JSON unmarshalling method for NodeCache.
// It converts the byte slices back to roaring bitmaps after JSON deserialization.
func (nc *NodeCache) UnmarshalJSON(data []byte) error {
	aux := &struct {
		AllParentsData  []byte `json:"allParentsData"`
		AllChildrenData []byte `json:"allChildrenData"`
		NodeID          uint32 `json:"ID"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal NodeCache data: %w", err)
	}
	nc.ID = aux.NodeID
	nc.AllParents = roaring.New()
	nc.AllChildren = roaring.New()
	if _, err := nc.AllParents.FromBuffer(aux.AllParentsData); err != nil {
		return fmt.Errorf("failed to convert AllParents data from buffer: %w", err)
	}
	if _, err := nc.AllChildren.FromBuffer(aux.AllChildrenData); err != nil {
		return fmt.Errorf("failed to convert AllChildren data from buffer: %w", err)
	}
	return nil
}

// MarshalJSON is a custom JSON marshalling tool.
// Roaring bitmaps can't be marshaled directly, so we need to call the roaring bitmaps function to convert the bitmaps to an []byte
// This takes the roaring bitmaps "Children" and "Parents" and converts them to byte slices called "ChildData" and "ParentData".
func (n *Node) MarshalJSON() ([]byte, error) {
	childData, err := n.Children.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert child bitmap to bytes: %w", err)
	}
	parentData, err := n.Parents.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert parent bitmap to bytes: %w", err)
	}
	return json.Marshal(&struct {
		Metadata   any    `json:"metadata"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
		ID         uint32 `json:"ID"`
	}{
		ID:         n.ID,
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
func (n *Node) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Metadata   any    `json:"metadata"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
		ID         uint32 `json:"ID"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal node data: %w", err)
	}
	n.ID = aux.ID
	n.Type = aux.Type
	n.Name = aux.Name
	n.Metadata = aux.Metadata
	n.Children = roaring.New()
	n.Parents = roaring.New()
	if _, err := n.Children.FromBuffer(aux.ChildData); err != nil {
		return fmt.Errorf("failed to convert child data from buffer: %w", err)
	}
	if _, err := n.Parents.FromBuffer(aux.ParentData); err != nil {
		return fmt.Errorf("failed to convert parent data from buffer: %w", err)
	}
	return nil
}

// AddNode becomes generic in terms of metadata
func AddNode(storage Storage, _type string, metadata any, name string) (*Node, error) {
	var ID uint32
	if id, err := storage.NameToID(name); err == nil {
		return storage.GetNode(id)
	} else {
		ID, err = storage.GenerateID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	n := &Node{
		ID:       ID,
		Type:     _type,
		Name:     name,
		Metadata: metadata,
		Children: roaring.New(),
		Parents:  roaring.New(),
	}
	nCache := &NodeCache{
		ID:          ID,
		AllParents:  roaring.New(),
		AllChildren: roaring.New(),
	}
	if err := storage.SaveNode(n); err != nil {
		return nil, fmt.Errorf("failed to save node: %w", err)
	}
	if err := storage.SaveCache(nCache); err != nil {
		return nil, err
	}
	return n, nil
}

// SetDependency now uses generic types for metadata
func (n *Node) SetDependency(storage Storage, neighbor *Node) error {
	if n == nil {
		return fmt.Errorf("cannot add dependency to nil node")
	}
	if neighbor == nil {
		return fmt.Errorf("cannot add dependency to nil node")
	}
	if n.ID == neighbor.ID {
		return ErrSelfDependency
	}
	if storage == nil {
		return fmt.Errorf("storages cannot be nil")
	}

	n.Children.Add(neighbor.ID)
	neighbor.Parents.Add(n.ID)

	if err := storage.SaveNode(n); err != nil {
		return fmt.Errorf("failed to save node: %w", err)
	}
	if err := storage.SaveNode(neighbor); err != nil {
		return fmt.Errorf("failed to save neighbor node: %w", err)
	}
	return nil
}

func (n *Node) queryBitmap(storage Storage, direction Direction) (*roaring.Bitmap, error) {
	if n == nil {
		return nil, fmt.Errorf("cannot query bitmap of nil node")
	}
	if storage == nil {
		return nil, fmt.Errorf("storages cannot be nil")
	}

	result := roaring.New()
	visited := make(map[uint32]bool)
	queue := []*Node{n}

	for len(queue) > 0 {
		curNode := queue[0]
		queue = queue[1:]

		if visited[curNode.ID] {
			continue
		}
		visited[curNode.ID] = true

		var bitmap *roaring.Bitmap
		switch direction {
		case ChildrenDirection:
			bitmap = curNode.Children
		case ParentsDirection:
			bitmap = curNode.Parents
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

	result.Add(n.ID)

	return result, nil
}

func (n *Node) QueryDependentsNoCache(storage Storage) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, ParentsDirection)
}

func (n *Node) QueryDependenciesNoCache(storage Storage) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, ChildrenDirection)
}

// QueryDependents checks if all nodes are cached, if so find the dependents in the cache, if not find the dependents without searching the cache
func (n *Node) QueryDependents(storage Storage) (*roaring.Bitmap, error) {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return nil, err
	}
	if len(uncachedNodes) > 0 {
		return n.QueryDependentsNoCache(storage)
	}

	nCache, err := storage.GetCache(n.ID)
	if err != nil {
		return nil, err
	}

	return nCache.AllParents, nil
}

func (n *Node) QueryDependencies(storage Storage) (*roaring.Bitmap, error) {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return nil, err
	}
	if len(uncachedNodes) > 0 {
		return n.QueryDependenciesNoCache(storage)
	}

	nCache, err := storage.GetCache(n.ID)
	if err != nil {
		return nil, err
	}

	return nCache.AllChildren, nil
}
