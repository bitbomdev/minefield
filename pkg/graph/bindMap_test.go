package graph

import (
	"strings"
	"testing"

	"github.com/RoaringBitmap/roaring"
)

func TestNewDataStore(t *testing.T) {
	ds := NewDataStore()
	if ds == nil {
		t.Error("expected non-nil DataStore")
	}
	if ds.data == nil {
		t.Error("expected non-nil data map")
	}
	if len(ds.data) != 0 {
		t.Error("expected empty data map")
	}
}

func TestNewNativeKeyManagement(t *testing.T) {
	nkm := NewNativeKeyManagement()
	if nkm == nil {
		t.Error("expected non-nil NativeKeyManagement")
	}
	if nkm.keyMap == nil {
		t.Error("expected non-nil keyMap")
	}
	if nkm.store == nil {
		t.Error("expected non-nil store")
	}
	if len(nkm.keyMap) != 0 {
		t.Error("expected empty keyMap")
	}
	if len(nkm.store.data) != 0 {
		t.Error("expected empty store data")
	}
}

func TestBindKeys(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		wantErr bool
	}{
		{
			name: "bind single key",
			keys: []string{"key1"},
		},
		{
			name: "bind multiple keys",
			keys: []string{"key1", "key2", "key3"},
		},
		{
			name: "bind empty keys",
			keys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nkm := NewNativeKeyManagement()
			hashKey, err := nkm.BindKeys(tt.keys)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the hash key has the expected prefix
				if !strings.HasPrefix(hashKey, "shared:") {
					t.Errorf("BindKeys() got hash key without 'shared:' prefix: %v", hashKey)
				}

				// Verify all keys are mapped to the same hash key
				for _, key := range tt.keys {
					mappedKey, exists := nkm.keyMap[key]
					if !exists {
						t.Errorf("key %v not found in keyMap", key)
						continue
					}
					if mappedKey != hashKey {
						t.Errorf("key %v mapped to %v, want %v", key, mappedKey, hashKey)
					}
				}
			}
		})
	}
}

func TestSetAndGet(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   *roaring.Bitmap
		wantErr bool
	}{
		{
			name:  "set and get single value",
			key:   "key1",
			value: roaring.BitmapOf(1, 2, 3),
		},
		{
			name:    "get non-existent key",
			key:     "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nkm := NewNativeKeyManagement()

			if tt.value != nil {
				err := nkm.Set(tt.key, *tt.value)
				if err != nil {
					t.Errorf("Set() error = %v", err)
					return
				}
			}

			got, err := nkm.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !got.Equals(tt.value) {
					t.Errorf("Get() = %v, want %v", got, tt.value)
				}
			}
		})
	}
}

func TestGetAllKeysAndValues(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string        // keys to bind
		setValue *roaring.Bitmap // value to set (nil if no value should be set)
		wantKeys []string
		wantVals []*roaring.Bitmap
		wantErr  bool
	}{
		{
			name:     "empty store",
			keys:     []string{},
			wantKeys: []string{},
			wantVals: []*roaring.Bitmap{},
		},
		{
			name:     "single bound key pair",
			keys:     []string{"key1", "key2"},
			setValue: roaring.BitmapOf(1, 2, 3),
			wantKeys: []string{"key1", "key2"},
			wantVals: []*roaring.Bitmap{
				roaring.BitmapOf(1, 2, 3),
				roaring.BitmapOf(1, 2, 3),
			},
		},
		{
			name:     "multiple bound keys",
			keys:     []string{"key1", "key2", "key3"},
			setValue: roaring.BitmapOf(1, 2, 3),
			wantKeys: []string{"key1", "key2", "key3"},
			wantVals: []*roaring.Bitmap{
				roaring.BitmapOf(1, 2, 3),
				roaring.BitmapOf(1, 2, 3),
				roaring.BitmapOf(1, 2, 3),
			},
		},
		{
			name:    "missing value for bound key",
			keys:    []string{"key1", "key2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nkm := NewNativeKeyManagement()

			// Bind keys if any
			if len(tt.keys) > 0 {
				_, err := nkm.BindKeys(tt.keys)
				if err != nil {
					t.Fatalf("Failed to bind keys: %v", err)
				}
			}

			// Set value if provided
			if tt.setValue != nil {
				err := nkm.Set(tt.keys[0], *tt.setValue)
				if err != nil {
					t.Fatalf("Failed to set value: %v", err)
				}
			}

			gotKeys, gotVals, err := nkm.GetAllKeysAndValues()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllKeysAndValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check keys length matches values length
			if len(gotKeys) != len(gotVals) {
				t.Errorf("GetAllKeysAndValues() returned mismatched lengths: keys=%d, values=%d",
					len(gotKeys), len(gotVals))
				return
			}

			// Check keys length matches expected
			if len(gotKeys) != len(tt.wantKeys) {
				t.Errorf("GetAllKeysAndValues() got %d keys, want %d",
					len(gotKeys), len(tt.wantKeys))
				return
			}

			// Create a map of got keys to values for easier comparison
			gotMap := make(map[string]*roaring.Bitmap)
			for i, key := range gotKeys {
				gotMap[key] = &gotVals[i]
			}

			// Check each expected key and value
			for i, wantKey := range tt.wantKeys {
				gotVal, exists := gotMap[wantKey]
				if !exists {
					t.Errorf("GetAllKeysAndValues() missing key %q", wantKey)
					continue
				}
				if !gotVal.Equals(tt.wantVals[i]) {
					t.Errorf("GetAllKeysAndValues() for key %q = %v, want %v",
						wantKey, gotVal, tt.wantVals[i])
				}
			}
		})
	}
}
