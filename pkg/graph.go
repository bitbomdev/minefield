package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/RoaringBitmap/roaring"
)

var ErrNodeAlreadyExists = errors.New("node with name already exists")

// Generic Node structure with metadata as generic type
type Node[T any] struct {
	Id         uint32          `json:"Id"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Metadata   T               `json:"metadata"`
	Child      *roaring.Bitmap `json:"child"`
	Parent     *roaring.Bitmap `json:"parent"`
	ChildData  []byte          `json:"childData"`
	ParentData []byte          `json:"parentData"`
}

type NodeCache struct {
	nodeID      uint32
	allParents  *roaring.Bitmap
	allChildren *roaring.Bitmap
}

func (n *Node[T]) GetID() uint32 { return n.Id }

func (n *Node[T]) GetChildren() *roaring.Bitmap { return n.Child }

func (n *Node[T]) GetParents() *roaring.Bitmap { return n.Parent }

func NewNodeCache(id uint32, allParents, allChildren *roaring.Bitmap) *NodeCache {
	return &NodeCache{
		nodeID:      id,
		allParents:  allParents,
		allChildren: allChildren,
	}
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
		Name       string `json:"name"`
		Metadata   T      `json:"metadata"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
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
		ID         uint32 `json:"Id"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		Metadata   T      `json:"metadata"`
		ChildData  []byte `json:"childData"`
		ParentData []byte `json:"parentData"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	n.Id = aux.ID
	n.Type = aux.Type
	n.Name = aux.Name
	n.Metadata = aux.Metadata
	n.Child = roaring.New()
	n.Parent = roaring.New()
	if _, err := n.Child.FromBuffer(aux.ChildData); err != nil {
		return err
	}
	if _, err := n.Parent.FromBuffer(aux.ParentData); err != nil {
		return err
	}
	return nil
}

// AddNode becomes generic in terms of metadata
func AddNode[T any](storage Storage[T], _type string, metadata T, name string) (*Node[T], error) {
	var ID uint32
	if id, err := storage.NameToID(name); err == nil {
		return storage.GetNode(id)
	} else {
		ID, err = storage.GenerateID()
		if err != nil {
			return nil, err
		}
	}

	n := &Node[T]{
		Id:       ID,
		Type:     _type,
		Name:     name,
		Metadata: metadata,
		Child:    roaring.New(),
		Parent:   roaring.New(),
	}
	nCache := &NodeCache{
		nodeID:      ID,
		allParents:  roaring.New(),
		allChildren: roaring.New(),
	}
	if err := storage.SaveNode(n); err != nil {
		return nil, err
	}
	if err := storage.SaveCache(nCache); err != nil {
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
	if err := storage.AddNodeToCachedStack(n.Id); err != nil {
		return err
	}
	if err := storage.AddNodeToCachedStack(neighbor.Id); err != nil {
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
			bitmap = curNode.Child
		case "parent":
			bitmap = curNode.Parent
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

	result.Remove(n.Id)

	return result, nil
}

func (n *Node[T]) QueryDependentsNoCache(storage Storage[T]) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, "parent")
}

func (n *Node[T]) QueryDependenciesNoCache(storage Storage[T]) (*roaring.Bitmap, error) {
	return n.queryBitmap(storage, "child")
}

// QueryDependents checks if all nodes are cached, if so find the dependents in the cache, if not find the dependents without searching the cache
func (n *Node[T]) QueryDependents(storage Storage[T]) (*roaring.Bitmap, error) {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return nil, err
	}
	if len(uncachedNodes) > 0 {
		return n.QueryDependentsNoCache(storage)
	}

	nCache, err := storage.GetCache(n.Id)
	if err != nil {
		return nil, err
	}

	return nCache.allParents, nil
}

func (n *Node[T]) QueryDependencies(storage Storage[T]) (*roaring.Bitmap, error) {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return nil, err
	}
	if len(uncachedNodes) > 0 {
		return n.QueryDependenciesNoCache(storage)
	}

	nCache, err := storage.GetCache(n.Id)
	if err != nil {
		return nil, err
	}

	return nCache.allChildren, nil
}

func GenerateDOT[T any](storage Storage[T]) (string, error) {
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
		for _, childID := range node.Child.ToArray() {
			dotBuilder.WriteString(fmt.Sprintf("%d -> %d;\n", node.GetID(), childID))
		}
	}
	dotBuilder.WriteString("}\n")
	return dotBuilder.String(), nil
}

func RenderGraph[T any](storage Storage[T]) error {
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
