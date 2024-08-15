package storage

import "github.com/bit-bom/minefield/pkg/graph"

// Storage is the interface that wraps the methods for a storage backend.
type Storage interface {
	NameToID(name string) (uint32, error)
	SaveNode(node *graph.Node) error
	GetNode(id uint32) (*graph.Node, error)
	GetNodes(ids []uint32) (map[uint32]*graph.Node, error)
	GetAllKeys() ([]uint32, error)
	SaveCache(cache *graph.NodeCache) error
	SaveCaches(cache []*graph.NodeCache) error
	ToBeCached() ([]uint32, error)
	AddNodeToCachedStack(id uint32) error
	GetCache(id uint32) (*graph.NodeCache, error)
	ClearCacheStack() error
	GenerateID() (uint32, error)
}
