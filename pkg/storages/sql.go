package storages

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bitbomdev/minefield/pkg/graph"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const key = "key = ?"
const KeyLike = "key LIKE ?"
const KeyIN = "key IN ?"

// KVStore represents the key-value storage table.
type KVStore struct {
	Key       string    `gorm:"primaryKey;uniqueIndex"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type CacheStack struct {
	ID        uint32    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type GlobalCounter struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// SQLStorage represents the storage backed by a SQL database.
type SQLStorage struct {
	DB *gorm.DB
}

// NewSQLStorage initializes a new SQLStorage with a SQLite database.
func NewSQLStorage(dsn string, useInMemory bool) (*SQLStorage, error) {
	var db *gorm.DB
	var err error
	if useInMemory {
		db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	} else {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQLDB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	storage := &SQLStorage{DB: db}

	if err := storage.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// Migrate performs the database migrations for SQLStorage.
func (s *SQLStorage) Migrate() error {
	return s.DB.AutoMigrate(&KVStore{}, &CacheStack{}, &GlobalCounter{})
}

// NameToID converts a node name to its corresponding ID.
func (s *SQLStorage) NameToID(name string) (uint32, error) {
	var kv KVStore
	if err := s.DB.First(&kv, key, NameToIDKey+name).Error; err != nil {
		return 0, fmt.Errorf("failed to get name-to-ID mapping: %w", err)
	}
	id, err := strconv.ParseUint(kv.Value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ID to integer: %w", err)
	}
	if id > math.MaxUint32 {
		return 0, fmt.Errorf("ID exceeds uint32 maximum value")
	}
	return uint32(id), nil
}

// SaveNode saves a node to the SQLite storage, handling node data, name-to-ID mapping, and caching.
func (s *SQLStorage) SaveNode(node *graph.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	// Marshal the node to JSON
	data, err := node.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	// Define keys
	nodeKey := fmt.Sprintf("%s%d", NodeKeyPrefix, node.ID)
	nameToIDKey := fmt.Sprintf("%s%s", NameToIDKey, node.Name)

	// Start a transaction to ensure atomicity
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Save the node data
		kvNode := KVStore{
			Key:   nodeKey,
			Value: string(data),
		}
		if err := tx.Save(&kvNode).Error; err != nil {
			return fmt.Errorf("failed to save node data: %w", err)
		}

		// Save the name-to-ID mapping
		kvMapping := KVStore{
			Key:   nameToIDKey,
			Value: fmt.Sprintf("%d", node.ID),
		}
		if err := tx.Save(&kvMapping).Error; err != nil {
			return fmt.Errorf("failed to save name-to-ID mapping: %w", err)
		}

		// Add the node ID to the cache stack without updating CreatedAt
		cacheEntry := CacheStack{
			ID: node.ID,
		}
		// There can be duplicates in the cache stack, so we use the clause.OnConflict
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}}, // Conflict on ID
			DoNothing: true,                          // Do nothing on conflict
		}).Create(&cacheEntry).Error; err != nil {
			return fmt.Errorf("failed to add node ID to cache stack: %w", err)
		}

		return nil
	})
}

// GetNode retrieves a node by its ID from the SQLite storage.
func (s *SQLStorage) GetNode(id uint32) (*graph.Node, error) {
	nodeKey := fmt.Sprintf("%s%d", NodeKeyPrefix, id)

	var kvNode KVStore
	if err := s.DB.First(&kvNode, "key = ?", nodeKey).Error; err != nil {
		return nil, fmt.Errorf("failed to get node data: %w", err)
	}

	var node graph.Node
	if err := node.UnmarshalJSON([]byte(kvNode.Value)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
	}

	return &node, nil
}

// GetNodes retrieves multiple nodes by their IDs.
func (s *SQLStorage) GetNodes(ids []uint32) (map[uint32]*graph.Node, error) {
	nodeKeys := generateNodeKeys(ids)
	var kvNodes []KVStore
	if err := s.DB.Where(KeyIN, nodeKeys).Find(&kvNodes).Error; err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	nodes := make(map[uint32]*graph.Node)
	for _, kvNode := range kvNodes {
		var node graph.Node
		if err := node.UnmarshalJSON([]byte(kvNode.Value)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node: %w", err)
		}
		idStr := strings.TrimPrefix(kvNode.Key, NodeKeyPrefix)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node ID: %w", err)
		}
		nodes[uint32(id)] = &node
	}
	return nodes, nil
}

// GetNodesByGlob retrieves nodes matching a glob pattern using SQL queries.
func (s *SQLStorage) GetNodesByGlob(pattern string) ([]*graph.Node, error) {
	// Construct the SQL LIKE pattern
	sqlPattern := convertGlobToSQLPattern(pattern)

	// Retrieve all name-to-ID mappings that match the pattern
	var mappings []KVStore
	if err := s.DB.Where(KeyLike, fmt.Sprintf("%s%s", NameToIDKey, sqlPattern)).Find(&mappings).Error; err != nil {
		return nil, fmt.Errorf("failed to get name-to-ID mappings with pattern %s: %w", pattern, err)
	}

	if len(mappings) == 0 {
		return []*graph.Node{}, nil // No matches found
	}

	// Extract IDs from the mappings
	ids := make([]uint32, 0, len(mappings))
	for _, mapping := range mappings {
		id, err := strconv.ParseUint(mapping.Value, 10, 32)
		if id > math.MaxUint32 {
			return nil, fmt.Errorf("ID exceeds uint32 maximum value")
		}
		if err != nil {
			return nil, fmt.Errorf("invalid ID format for key %s: %w", mapping.Key, err)
		}
		ids = append(ids, uint32(id))
	}
	var nodes []KVStore
	var resultNodes []*graph.Node
	// Retrieve nodes with the extracted IDs
	if err := s.DB.Where("key IN ?", generateNodeKeys(ids)).Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve nodes for IDs %v: %w", ids, err)
	}
	for _, node := range nodes {
		var graphNode graph.Node
		if err := graphNode.UnmarshalJSON([]byte(node.Value)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
		}
		resultNodes = append(resultNodes, &graphNode)
	}

	return resultNodes, nil
}

// generateNodeKeys creates a slice of node keys based on IDs.
func generateNodeKeys(ids []uint32) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("%s%d", NodeKeyPrefix, id)
	}
	return keys
}

// GetAllKeys retrieves all node IDs.
func (s *SQLStorage) GetAllKeys() ([]uint32, error) {
	var kvNodes []KVStore
	if err := s.DB.Where(KeyLike, NodeKeyPrefix+"%").Find(&kvNodes).Error; err != nil {
		return nil, fmt.Errorf("failed to get all node IDs: %w", err)
	}
	ids := make([]uint32, len(kvNodes))
	for i, kvNode := range kvNodes {
		var graphNode graph.Node
		if err := graphNode.UnmarshalJSON([]byte(kvNode.Value)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
		}
		ids[i] = graphNode.ID
	}
	return ids, nil
}

// SaveCache saves a node cache.
func (s *SQLStorage) SaveCache(cache *graph.NodeCache) error {
	cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, cache.ID)
	data, err := cache.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	kvCache := KVStore{
		Key:   cacheKey,
		Value: string(data),
	}
	if err := s.DB.Save(&kvCache).Error; err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}
	return nil
}

// SaveCaches saves multiple node caches.
func (s *SQLStorage) SaveCaches(caches []*graph.NodeCache) error {
	kvCaches := make([]KVStore, len(caches))
	for i, cache := range caches {
		cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, cache.ID)
		data, err := cache.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal cache: %w", err)
		}
		kvCaches[i] = KVStore{
			Key:   cacheKey,
			Value: string(data),
		}
	}
	if err := s.DB.Create(&kvCaches).Error; err != nil {
		return fmt.Errorf("failed to save caches: %w", err)
	}
	return nil
}

// RemoveAllCaches removes all caches from the database.
func (s *SQLStorage) RemoveAllCaches() error {
	if err := s.DB.Delete(&KVStore{}, "key LIKE ?", CacheKeyPrefix+"%").Error; err != nil {
		return fmt.Errorf("failed to remove all caches: %w", err)
	}
	return nil
}

// ToBeCached retrieves IDs of nodes to be cached.
func (s *SQLStorage) ToBeCached() ([]uint32, error) {
	var cacheStack []CacheStack
	if err := s.DB.Find(&cacheStack).Error; err != nil {
		return nil, fmt.Errorf("failed to get cache stack: %w", err)
	}
	ids := make([]uint32, len(cacheStack))
	for i, cache := range cacheStack {
		ids[i] = cache.ID
	}
	return ids, nil
}

// AddNodeToCachedStack adds a node ID to the cached stack.
func (s *SQLStorage) AddNodeToCachedStack(id uint32) error {
	cacheEntry := CacheStack{
		ID: id,
	}
	if err := s.DB.Create(&cacheEntry).Error; err != nil {
		return fmt.Errorf("failed to add node ID to cache stack: %w", err)
	}
	return nil
}

// GetCache retrieves a cache by its ID.
func (s *SQLStorage) GetCache(id uint32) (*graph.NodeCache, error) {
	cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, id)
	var kvCache KVStore
	if err := s.DB.First(&kvCache, key, cacheKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if the cache does not exist
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	var cache graph.NodeCache
	if err := cache.UnmarshalJSON([]byte(kvCache.Value)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}
	return &cache, nil
}

// GetCaches retrieves multiple caches by their IDs.
func (s *SQLStorage) GetCaches(ids []uint32) (map[uint32]*graph.NodeCache, error) {
	cacheKeys := make([]string, len(ids))
	for i, id := range ids {
		cacheKeys[i] = fmt.Sprintf("%s%d", CacheKeyPrefix, id)
	}
	var kvCaches []KVStore
	if err := s.DB.Where(KeyIN, cacheKeys).Find(&kvCaches).Error; err != nil {
		return nil, fmt.Errorf("failed to get caches: %w", err)
	}
	caches := make(map[uint32]*graph.NodeCache, len(kvCaches))
	for _, kvCache := range kvCaches {
		var cache graph.NodeCache
		if err := cache.UnmarshalJSON([]byte(kvCache.Value)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
		}
		idStr := strings.TrimPrefix(kvCache.Key, CacheKeyPrefix)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ID from key: %w", err)
		}
		if id > math.MaxUint32 {
			return nil, fmt.Errorf("ID exceeds uint32 maximum value")
		}
		caches[uint32(id)] = &cache
	}
	return caches, nil
}

// ClearCacheStack clears the cache stack.
func (s *SQLStorage) ClearCacheStack() error {
	if err := s.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&CacheStack{}).Error; err != nil {
		return fmt.Errorf("failed to clear cache stack: %w", err)
	}
	return nil
}

// GenerateID generates a new unique ID by inserting a new GlobalCounter and retrieving its ID.
func (s *SQLStorage) GenerateID() (uint32, error) {
	counter := GlobalCounter{}
	if err := s.DB.Create(&counter).Error; err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	return counter.ID, nil
}

// GetCustomData retrieves custom data based on tag and key.
func (s *SQLStorage) GetCustomData(tag, key string) (map[string][]byte, error) {
	// TODO: Implement using GORM
	return nil, fmt.Errorf("not implemented")
}

// AddOrUpdateCustomData adds or updates custom data based on tag, key, and data key.
func (s *SQLStorage) AddOrUpdateCustomData(tag, key string, dataKey string, data []byte) error {
	// TODO: Implement using GORM
	return fmt.Errorf("not implemented")
}

// convertGlobToSQLPattern converts a glob pattern to a SQL LIKE pattern.
// It replaces '*' with '%' and '?' with '_'. It also escapes existing '%' and '_' characters.
func convertGlobToSQLPattern(pattern string) string {
	var sb strings.Builder
	for _, char := range pattern {
		switch char {
		case '*':
			sb.WriteByte('%')
		case '?':
			sb.WriteByte('_')
		case '%', '_':
			sb.WriteByte('\\')
			sb.WriteRune(char)
		default:
			sb.WriteRune(char)
		}
	}
	return sb.String()
}
