package pkg

import (
	"sync"

	"github.com/RoaringBitmap/roaring"
)

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
	storageInstance Storage[any]
	once            sync.Once
)

func GetStorageInstance(addr string) Storage[any] {
	once.Do(func() {
		storageInstance = NewRedisStorage[any](addr)
	})
	return storageInstance
}
