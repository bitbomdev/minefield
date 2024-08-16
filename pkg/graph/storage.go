package graph

// Storage is the interface that wraps the methods for a storage backend.
type Storage interface {
	NameToID(name string) (uint32, error)
	SaveNode(node *Node) error
	GetNode(id uint32) (*Node, error)
	GetNodes(ids []uint32) (map[uint32]*Node, error)
	GetAllKeys() ([]uint32, error)
	SaveCache(cache *NodeCache) error
	SaveCaches(cache []*NodeCache) error
	ToBeCached() ([]uint32, error)
	AddNodeToCachedStack(id uint32) error
	GetCache(id uint32) (*NodeCache, error)
	ClearCacheStack() error
	GenerateID() (uint32, error)
}
