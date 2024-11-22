package helpers

// ellipsis is the string prepended to truncated strings
const ellipsis = "..."

// TruncateString truncates a string to the specified maximum length.
// If the string length exceeds maxLength, it will be truncated and
// prepended with "...". If maxLength is less than 4 (minimum length
// needed for "..." + 1 character), the original string is returned.
//
// Example:
//   TruncateString("hello world", 8) // returns "...world"
//   TruncateString("hi", 8)          // returns "hi"
func TruncateString(str string, maxLength int) string {
	if maxLength < 4 {
		return str
	}
	if len(str) <= maxLength {
		return str
	}
	return ellipsis + str[len(str)-maxLength+len(ellipsis):]
}
