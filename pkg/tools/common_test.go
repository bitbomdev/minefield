package tools

import (
	"testing"
)

// TestSanitizeFilename tests the SanitizeFilename function.
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "file/name", expected: "file_name"},
		{input: "file\\name", expected: "file_name"},
		{input: "file:name", expected: "file_name"},
		{input: "file*name", expected: "file_name"},
		{input: "file?name", expected: "file_name"},
		{input: "file\"name", expected: "file_name"},
		{input: "file<name", expected: "file_name"},
		{input: "file>name", expected: "file_name"},
		{input: "file|name", expected: "file_name"},
		{input: "file/name\\name:name*name?name\"name<name>name|name", expected: "file_name_name_name_name_name_name_name_name_name"},
		{input: "filename", expected: "filename"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := SanitizeFilename(test.input)
			if result != test.expected {
				t.Errorf("SanitizeFilename(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}

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
