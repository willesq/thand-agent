package common

import "strings"

// Helper function to check if a string contains a substring (case-insensitive)
func ContainsInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
