package tools

import (
	"strings"
)

// SanitizeFilename replaces characters that are not allowed in filenames with underscores.
func SanitizeFilename(filename string) string {
	// Define a set of characters that are not allowed in filenames.
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}

	// Replace each invalid character with an underscore.
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	return filename
}

func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return "..." + str[len(str)-maxLength+3:]
}
