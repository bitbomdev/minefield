package pkg

import (
	"sync"

	"github.com/RoaringBitmap/roaring"
)

type Storage[T any] interface {
	SaveNode(node *Node[T]) error
	GetNode(id uint32) (*Node[T], error)
	GetAllKeys() ([]uint32, error)
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
