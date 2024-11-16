package storages

import (
	"fmt"

	"github.com/bitbomdev/minefield/pkg/graph"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLStorage is the SQLite implementation of the Storage interface using GORM.
type SQLStorage struct {
	DB *gorm.DB
}

// NewSQLiteStorage creates a new instance of SQLStorage with the given Data Source Name (DSN).
func NewSQLiteStorage(dsn string) (*SQLStorage, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	return &SQLStorage{DB: db}, nil
}

// NameToID converts a node name to its corresponding ID.
func (s *SQLStorage) NameToID(name string) (uint32, error) {
	// TODO: Implement using GORM
	return 0, nil
}

// SaveNode saves a node to the database.
func (s *SQLStorage) SaveNode(node *graph.Node) error {
	// TODO: Implement using GORM
	return nil
}

// GetNode retrieves a node by its ID.
func (s *SQLStorage) GetNode(id uint32) (*graph.Node, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// GetNodes retrieves multiple nodes by their IDs.
func (s *SQLStorage) GetNodes(ids []uint32) (map[uint32]*graph.Node, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// GetNodesByGlob retrieves nodes matching a glob pattern.
func (s *SQLStorage) GetNodesByGlob(pattern string) ([]*graph.Node, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// GetAllKeys retrieves all node IDs.
func (s *SQLStorage) GetAllKeys() ([]uint32, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// SaveCache saves a node cache.
func (s *SQLStorage) SaveCache(cache *graph.NodeCache) error {
	// TODO: Implement using GORM
	return nil
}

// SaveCaches saves multiple node caches.
func (s *SQLStorage) SaveCaches(caches []*graph.NodeCache) error {
	// TODO: Implement using GORM
	return nil
}

// RemoveAllCaches removes all caches from the database.
func (s *SQLStorage) RemoveAllCaches() error {
	// TODO: Implement using GORM
	return nil
}

// ToBeCached retrieves IDs of nodes to be cached.
func (s *SQLStorage) ToBeCached() ([]uint32, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// AddNodeToCachedStack adds a node ID to the cached stack.
func (s *SQLStorage) AddNodeToCachedStack(id uint32) error {
	// TODO: Implement using GORM
	return nil
}

// GetCache retrieves a cache by its ID.
func (s *SQLStorage) GetCache(id uint32) (*graph.NodeCache, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// GetCaches retrieves multiple caches by their IDs.
func (s *SQLStorage) GetCaches(ids []uint32) (map[uint32]*graph.NodeCache, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// ClearCacheStack clears the cache stack.
func (s *SQLStorage) ClearCacheStack() error {
	// TODO: Implement using GORM
	return nil
}

// GenerateID generates a new unique ID.
func (s *SQLStorage) GenerateID() (uint32, error) {
	// TODO: Implement using GORM
	return 0, nil
}

// GetCustomData retrieves custom data based on tag and key.
func (s *SQLStorage) GetCustomData(tag, key string) (map[string][]byte, error) {
	// TODO: Implement using GORM
	return nil, nil
}

// AddOrUpdateCustomData adds or updates custom data based on tag, key, and data key.
func (s *SQLStorage) AddOrUpdateCustomData(tag, key string, dataKey string, data []byte) error {
	// TODO: Implement using GORM
	return nil
}
