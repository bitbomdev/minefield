package pkg

import (
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/bit-bom/bitbom/pkg/ingest"
)

func TestParseAndExecute(t *testing.T) {
	storage := GetStorageInstance("localhost:6379")
	err := ingest.SBOM("../test", storage)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		script  string
		want    *roaring.Bitmap
		wantErr bool
	}{
		{
			name:   "Simple dependents query",
			script: "dependents PACKAGE pkg:generic/dep1@1.0.0",
			want:   roaring.BitmapOf(3, 4),
		},
		{
			name:   "Simple dependencies query",
			script: "dependencies PACKAGE pkg:generic/lib-A@1.0.0",
			want:   roaring.BitmapOf(1, 2),
		},
		{
			name:    "Invalid token",
			script:  "invalid PACKAGE pkg:generic/lib-A@1.0.0",
			wantErr: true,
		},
		{
			name:   "Combine dependents and dependencies with OR",
			script: "dependents PACKAGE pkg:generic/dep1@1.0.0 or dependencies PACKAGE pkg:generic/dep1@1.0.0",
			want:   roaring.BitmapOf(2, 3, 4),
		},
		{
			name:    "Mismatched parentheses",
			script:  "dependents PACKAGE pkg:generic/lib-A@1.0.0 ]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndExecute[any](tt.script, storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAndExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equals(tt.want) {
				t.Errorf("ParseAndExecute() got = %v, want %v", result, tt.want)
			}
		})
	}
}
