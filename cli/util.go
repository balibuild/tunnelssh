package cli

import "strings"

// IsTrue fun
func IsTrue(s string) bool {
	lower := strings.ToLower(s)
	return lower == "true" || lower == "yes" || lower == "on" || lower == "1"
}
