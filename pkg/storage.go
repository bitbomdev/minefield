package pkg

import (
	"sync"

	"github.com/RoaringBitmap/roaring"
)

// Storage is the interface that wraps the methods for a storage backend.
type Storage[T any] interface {
	NameToID(name string) (uint32, error)
	IDToName(id uint32) (string, error)
	SaveNode(node *Node[T]) error
	GetNode(id uint32) (*Node[T], error)
	GetAllKeys() ([]uint32, error)
	SaveCache(cache *NodeCache) error
	ToBeCached() ([]uint32, error)
	AddNodeToCachedStack(id uint32) error
	GetCache(id uint32) (*NodeCache, error)
	ClearCacheStack() error
	SetDependency(nodeID, neighborID uint32) error
	QueryDependents(nodeID uint32) (*roaring.Bitmap, error)
	QueryDependencies(nodeID uint32) (*roaring.Bitmap, error)
	GenerateID() (uint32, error)
}

var (
	// storageInstance is the singleton instance of the storage interface.
	storageInstance Storage[any]
	// once is a sync.Once that ensures that the storage instance is only initialized once.
	once sync.Once
)

// GetStorageInstance returns a singleton instance of the storage interface.
func GetStorageInstance(addr string) Storage[any] {
	once.Do(func() {
		storageInstance = NewRedisStorage[any](addr)
	})
	return storageInstance
}
