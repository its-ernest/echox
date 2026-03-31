package abuse

import "strings"

// matchPath handles basic prefix and wildcard matching for abuse rules.
// It supports exact matches (e.g., "/.env") and prefix wildcards (e.g., "/admin*").
func matchPath(pattern, path string) bool {
	// 1. Exact match
	if pattern == path {
		return true
	}

	// 2. Wildcard match (e.g., /wp-admin*)
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}
