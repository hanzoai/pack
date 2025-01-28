package resolver

import (
	"testing"

	"github.com/railwayapp/railpack/core/mise"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesWithDefaults(t *testing.T) {
	pkg := NewRequestedPackage("node", "18")
	assert.Equal(t, "18", pkg.Version)
	assert.Equal(t, DefaultSource, pkg.Source)

	pkg.SetVersion("22", "package.json engines")
	assert.Equal(t, "22", pkg.Version)
	assert.Equal(t, "package.json engines", pkg.Source)
}

func TestPackageResolver(t *testing.T) {
	resolver, err := NewResolver(mise.TestInstallDir)
	require.NoError(t, err)

	// Set up Node.js
	node := resolver.Default("node", "18")
	resolver.Version(node, "23", "package.json engines")

	// Set up Bun
	resolver.Default("bun", "latest")

	// Set up Python
	python := resolver.Default("python", "3.11")
	resolver.Version(python, "3.12", "PYTHON_VERSION environment variable")
	resolver.Version(python, "3.13", ".python-version")

	// Resolve all packages
	resolvedPackages, err := resolver.ResolvePackages()
	require.NoError(t, err)
	assert.Equal(t, 3, len(resolvedPackages))

	// Check Node.js resolution
	nodeResolved := resolvedPackages["node"]
	require.NotNil(t, nodeResolved)
	require.NotNil(t, nodeResolved.ResolvedVersion)
	assert.Contains(t, *nodeResolved.ResolvedVersion, "23")

	// Check Bun resolution
	bunResolved := resolvedPackages["bun"]
	assert.NotNil(t, bunResolved)

	// Check Python resolution
	pythonResolved := resolvedPackages["python"]
	require.NotNil(t, pythonResolved)
	require.NotNil(t, pythonResolved.ResolvedVersion)
	assert.Contains(t, *pythonResolved.ResolvedVersion, "3.13")
}
