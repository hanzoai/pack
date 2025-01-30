package resolver

import "strings"

func resolveToFuzzyVersion(version string) string {
	// Remove any whitespace
	version = strings.TrimSpace(version)

	// Handle empty string and "*" cases
	if version == "" || version == "*" {
		return "latest"
	}

	// Handle range notation (e.g. ">=22 <23")
	if strings.Contains(version, ">=") || strings.Contains(version, "<") {
		parts := strings.Fields(version)
		// Take the first version number we find
		for _, part := range parts {
			if v := strings.TrimPrefix(strings.TrimPrefix(part, ">="), "<"); v != part {
				version = v
				break
			}
		}
	}

	// Remove any prefix characters (^, ~, v)
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, "v")

	// Replace .x with empty string (e.g. "14.x" -> "14")
	version = strings.ReplaceAll(version, ".x", "")

	// Remove any trailing dots
	version = strings.TrimRight(version, ".")

	return version
}
