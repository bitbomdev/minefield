package pkg

import (
	"sync"
)

// Storage is the interface that wraps the methods for a storage backend.
type Storage interface {
	NameToID(name string) (uint32, error)
	SaveNode(node *Node) error
	GetNode(id uint32) (*Node, error)
	GetNodes(ids []uint32) (map[uint32]*Node, error)
	GetAllKeys() ([]uint32, error)
	SaveCache(cache *NodeCache) error
	ToBeCached() ([]uint32, error)
	AddNodeToCachedStack(id uint32) error
	GetCache(id uint32) (*NodeCache, error)
	ClearCacheStack() error
	GenerateID() (uint32, error)
}

var (
	// storageInstance is the singleton instance of the storage interface.
	storageInstance Storage
	// once is a sync.Once that ensures that the storage instance is only initialized once.
	once sync.Once
)

// GetStorageInstance returns a singleton instance of the storage interface.
func GetStorageInstance(addr string) Storage {
	once.Do(func() {
		storageInstance = NewRedisStorage(addr)
	})
	return storageInstance
}
