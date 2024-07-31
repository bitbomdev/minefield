package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

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
	nodeID      uint32
	allParents  *roaring.Bitmap
	allChildren *roaring.Bitmap
}

func (n *Node) GetID() uint32 { return n.ID }

func (n *Node) GetChildren() *roaring.Bitmap { return n.Children }

func (n *Node) GetParents() *roaring.Bitmap { return n.Parents }

func NewNodeCache(id uint32, allParents, allChildren *roaring.Bitmap) *NodeCache {
	return &NodeCache{
		nodeID:      id,
		allParents:  allParents,
		allChildren: allChildren,
	}
}

// MarshalJSON is a custom JSON marshalling method for NodeCache.
// It converts the roaring bitmaps to byte slices for JSON serialization.
func (nc *NodeCache) MarshalJSON() ([]byte, error) {
	allParentsData, err := nc.allParents.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert allParents bitmap to bytes: %w", err)
	}
	allChildrenData, err := nc.allChildren.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert allChildren bitmap to bytes: %w", err)
	}
	return json.Marshal(&struct {
		NodeID          uint32 `json:"nodeID"`
		AllParentsData  []byte `json:"allParentsData"`
		AllChildrenData []byte `json:"allChildrenData"`
	}{
		NodeID:          nc.nodeID,
		AllParentsData:  allParentsData,
		AllChildrenData: allChildrenData,
	})
}

// UnmarshalJSON is a custom JSON unmarshalling method for NodeCache.
// It converts the byte slices back to roaring bitmaps after JSON deserialization.
func (nc *NodeCache) UnmarshalJSON(data []byte) error {
	aux := &struct {
		NodeID          uint32 `json:"nodeID"`
		AllParentsData  []byte `json:"allParentsData"`
		AllChildrenData []byte `json:"allChildrenData"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal NodeCache data: %w", err)
	}
	nc.nodeID = aux.NodeID
	nc.allParents = roaring.New()
	nc.allChildren = roaring.New()
	if _, err := nc.allParents.FromBuffer(aux.AllParentsData); err != nil {
		return fmt.Errorf("failed to convert allParents data from buffer: %w", err)
	}
	if _, err := nc.allChildren.FromBuffer(aux.AllChildrenData); err != nil {
		return fmt.Errorf("failed to convert allChildren data from buffer: %w", err)
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
		nodeID:      ID,
		allParents:  roaring.New(),
		allChildren: roaring.New(),
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
		return fmt.Errorf("storage cannot be nil")
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
		return nil, fmt.Errorf("storage cannot be nil")
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

	result.Remove(n.ID)

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

	return nCache.allParents, nil
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

	return nCache.allChildren, nil
}

func GenerateDOT(storage Storage) (string, error) {
	keys, err := storage.GetAllKeys()
	if err != nil {
		return "", err
	}

	var dotBuilder strings.Builder
	dotBuilder.WriteString("digraph G {\n")
	dotBuilder.WriteString("node [shape=ellipse, style=filled, fillcolor=lightblue];\n") // Node style
	dotBuilder.WriteString("edge [color=gray];\n")                                       // Edge style

	for _, key := range keys {
		node, err := storage.GetNode(key)
		if err != nil {
			return "", err
		}

		// Add the node with a label that includes type and additional metadata if needed
		label := fmt.Sprintf("%s\\nMetadata: %v", node.Type, node.Metadata)
		dotBuilder.WriteString(fmt.Sprintf("%d [label=\"%s\"];\n", node.GetID(), label))

		// Add edges for children
		for _, childID := range node.Children.ToArray() {
			dotBuilder.WriteString(fmt.Sprintf("%d -> %d;\n", node.GetID(), childID))
		}
	}
	dotBuilder.WriteString("}\n")
	return dotBuilder.String(), nil
}

func RenderGraph(storage Storage) error {
	dotString, err := GenerateDOT(storage)
	if err != nil {
		return err
	}

	cmd := exec.Command("dot", "-Tpng", "-o", "graph.png", "-Kfdp") // Using fdp for a spring model layout
	cmd.Stdin = strings.NewReader(dotString)
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Graph rendered as graph.png using fdp layout")
	return nil
}
