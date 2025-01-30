package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPreviousVersions(t *testing.T) {
	input := []string{
		"basic@1.0.0",
		"caret@^2.4",
		"tilde@~3.1.3",
		"vprefix@v4.0.0",
		"xnotation@14.x",
		"range@>=22 <23",
		"wildcard@*",
	}

	previousVersions := getPreviousVersions(input)

	expected := map[string]string{
		"basic":     "1.0.0",
		"caret":     "^2.4",
		"tilde":     "~3.1.3",
		"vprefix":   "v4.0.0",
		"xnotation": "14.x",
		"range":     ">=22 <23",
		"wildcard":  "*",
	}

	require.Equal(t, expected, previousVersions)
}
