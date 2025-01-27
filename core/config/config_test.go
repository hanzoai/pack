package config

import (
	"testing"

	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/stretchr/testify/require"
)

func TestEmptyConfig(t *testing.T) {
	config := EmptyConfig()
	require.NotNil(t, config)
	require.Empty(t, config.Caches)
	require.Empty(t, config.Packages)
	require.Empty(t, config.AptPackages)
	require.Empty(t, config.Steps)
}

func TestMergeConfig(t *testing.T) {
	config1 := &Config{
		BaseImage: "ubuntu:20.04",
		Packages: map[string]string{
			"python": "latest",
			"node":   "22",
		},
		AptPackages: []string{"git"},
		Start: plan.Start{
			Command: "python app.py",
		},
	}

	config2 := &Config{
		BaseImage: "ubuntu:22.04",
		Packages: map[string]string{
			"node": "23",
		},
		AptPackages: []string{"curl"},
		Start: plan.Start{
			Command: "node server.js",
		},
	}

	result := config1.Merge(config2)

	// Check that BaseImage takes the last value
	require.Equal(t, "ubuntu:22.04", result.BaseImage)

	// Check that packages are merged with last value winning
	require.Equal(t, "latest", result.Packages["python"])
	require.Equal(t, "23", result.Packages["node"])

	// Check that arrays are extended
	require.ElementsMatch(t, []string{"git", "curl"}, result.AptPackages)

	// Check that Start takes the last non-empty value
	require.Equal(t, "node server.js", result.Start.Command)

	// Test that when second config has empty Start, first config's Start is kept
	config2.Start = plan.Start{}
	result = config1.Merge(config2)
	require.Equal(t, "python app.py", result.Start.Command)
}
