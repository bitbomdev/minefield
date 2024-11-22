package helpers

func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return "..." + str[len(str)-maxLength+3:]
}
