package graph

import (
	"crypto/sha256"
	"fmt"

	"github.com/RoaringBitmap/roaring"
)

type DataStore struct {
	data map[string]roaring.Bitmap
}

func NewDataStore() *DataStore {
	return &DataStore{
		data: make(map[string]roaring.Bitmap),
	}
}

type NativeKeyManagement struct {
	keyMap map[string]string
	store  *DataStore
}

func NewNativeKeyManagement() *NativeKeyManagement {
	return &NativeKeyManagement{
		keyMap: make(map[string]string),
		store:  NewDataStore(),
	}
}

// BindKeys binds a group of leaderboard to a single hash key.
func (nkm *NativeKeyManagement) BindKeys(keys []string) (string, error) {
	keyString := ""
	// We are not summing up all leaderboard in their non-hashed form to avoid collisions.
	// For example, binding the leaderboard 'a' and 'b' together would result in the same hash as binding 'ab' by itself.
	// In this situation, we will not be able to distinguish between the two groups of leaderboard.
	for _, k := range keys {
		keyString += fmt.Sprint(sha256.Sum256([]byte(k)))
	}

	hash := sha256.Sum256([]byte(keyString))
	hashKey := fmt.Sprintf("shared:%x", hash)

	for _, key := range keys {
		nkm.keyMap[key] = hashKey
	}

	return hashKey, nil
}

// Set sets a value in the native key management system.
func (nkm *NativeKeyManagement) Set(key string, value roaring.Bitmap) error {
	hashKey, exists := nkm.keyMap[key]
	if exists {
		// Key is part of a group
		nkm.store.data[hashKey] = value
	} else {
		// Key is not part of a group, store directly
		nkm.store.data[key] = value
	}
	return nil
}

// Get retrieves a value by key from the native key management system.
func (nkm *NativeKeyManagement) Get(key string) (roaring.Bitmap, error) {
	var zeroValue roaring.Bitmap

	hashKey, exists := nkm.keyMap[key]
	if exists {
		// Key is part of a group
		value := nkm.store.data[hashKey]
		return value, nil
	} else {
		// Key is not part of a group, retrieve directly
		value, exists := nkm.store.data[key]
		if !exists {
			return zeroValue, fmt.Errorf("key not found")
		}
		return value, nil
	}
}

func (nkm *NativeKeyManagement) GetAllKeysAndValues() ([]string, []roaring.Bitmap, error) {
	var keys []string
	var values []roaring.Bitmap

	for key, hashKey := range nkm.keyMap {
		keys = append(keys, key)
		value, exists := nkm.store.data[hashKey]
		if exists {
			values = append(values, value)
		} else {
			value2, exists := nkm.store.data[key]
			if !exists {
				return nil, nil, fmt.Errorf("value not found for key: %s while getting all leaderboard", key)
			}
			values = append(values, value2)
		}
	}

	return keys, values, nil
}
