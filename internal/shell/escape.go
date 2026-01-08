package shell

import (
	"strings"
)

// ShellEscape escapes a string for safe use in shell commands
// This prevents command injection by properly quoting and escaping special characters
func ShellEscape(s string) string {
	// If string is already safe (alphanumeric, dash, underscore, dot, slash), return as-is
	if isSafeString(s) {
		return s
	}

	// Use single quotes and escape any single quotes in the string
	// Single quotes preserve everything literally except single quotes themselves
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}

// isSafeString checks if a string contains only safe characters
// Safe characters: alphanumeric, dash, underscore, dot, forward slash
func isSafeString(s string) bool {
	if s == "" {
		return false
	}

	// Path traversal sequences are never safe
	if strings.Contains(s, "..") {
		return false
	}

	for _, r := range s {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == '/') {
			return false
		}
	}
	return true
}

// ValidateVolumeName validates that a volume name is safe to use
// Volume names should only contain alphanumeric characters, dashes, underscores, and dots
func ValidateVolumeName(name string) bool {
	if name == "" || len(name) > 255 {
		return false
	}

	// Volume names must not start with - or .
	if name[0] == '-' || name[0] == '.' {
		return false
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}

	// Only allow alphanumeric, dash, underscore, dot
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.') {
			return false
		}
	}

	return true
}

// SanitizePathForRemote ensures a remote path is safe
// Prevents path traversal and ensures absolute paths
func SanitizePathForRemote(path string) string {
	// Remove any path traversal attempts
	path = strings.ReplaceAll(path, "..", "")

	// Ensure path is absolute (starts with /)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Remove multiple consecutive slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	return path
}
