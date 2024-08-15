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
