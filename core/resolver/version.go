package resolver

import "strings"

func resolveToFuzzyVersion(version string) string {
	// Remove any whitespace
	version = strings.TrimSpace(version)

	// Handle empty string and "*" cases
	if version == "" || version == "*" {
		return "latest"
	}

	// Handle range notation (e.g. ">=22 <23" or ">= 22" or ">=20.0.0")
	if strings.Contains(version, ">=") || strings.Contains(version, "<") {
		parts := strings.Fields(version)
		for i, part := range parts {
			if strings.HasPrefix(part, ">=") {
				// Version number is either after the >= in this part, or in the next part
				v := strings.TrimPrefix(part, ">=")
				if v == "" && i+1 < len(parts) {
					v = parts[i+1]
				}
				return strings.Split(strings.TrimSpace(v), ".")[0]
			}
		}
	}

	// Handle caret notation by only keeping major version
	if strings.HasPrefix(version, "^") {
		version = strings.TrimPrefix(version, "^")
		parts := strings.Split(version, ".")
		return parts[0]
	}

	// Remove any prefix characters (~, v)
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, "v")

	// Replace .x with empty string (e.g. "14.x" -> "14")
	version = strings.ReplaceAll(version, ".x", "")

	// Remove any trailing dots
	version = strings.TrimRight(version, ".")

	return version
}
