package helpers

import (
	"testing"
)

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		maxLength int
		want      string
	}{
		{
			name:      "Path shorter than maxLength",
			id:        "short/path",
			maxLength: 20,
			want:      "short/path",
		},
		{
			name:      "Path exactly maxLength",
			id:        "exactly20/characters",
			maxLength: 20,
			want:      "exactly20/characters",
		},
		{
			name:      "Path longer than maxLength",
			id:        "this/is/a/very/long/path/that/needs/truncation",
			maxLength: 20,
			want:      ".../needs/truncation",
		},
		{
			name:      "Path with special characters",
			id:        "special/!@#$%^&*()_+",
			maxLength: 10,
			want:      "...^&*()_+",
		},
		{
			name:      "Empty path",
			id:        "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "Very small maxLength",
			id:        "small/path",
			maxLength: 5,
			want:      "...th",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.id, tt.maxLength); got != tt.want {
				t.Errorf("truncatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
