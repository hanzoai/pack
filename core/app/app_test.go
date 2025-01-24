package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPackageJSON represents the package.json structure for testing
type PackageJSON struct {
	Name    string `json:"name"`
	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`
	Scripts struct {
		PostStart string
		PreStart  string
		Start     string
	}
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`

	AllScripts map[string]string `json:"scripts"`
}

func TestApp(t *testing.T) {
	app, err := NewApp("../../examples/node-bun")
	require.NoError(t, err)

	content, err := app.ReadFile("package.json")
	require.NoError(t, err)
	require.Contains(t, content, "node-bun")

	var packageJSON PackageJSON
	err = app.ReadJSON("package.json", &packageJSON)
	require.NoError(t, err)
	require.Equal(t, packageJSON.Name, "node-bun")

	files, err := app.FindFiles("*.ts")
	require.NoError(t, err)
	require.Equal(t, len(files), 1)
	require.Equal(t, files[0], "index.ts")
}

func TestAppAbsolutePath(t *testing.T) {
	relPath := "../../examples/node-bun"
	absPath, err := filepath.Abs(relPath)
	require.NoError(t, err)

	app, err := NewApp(absPath)
	require.NoError(t, err)

	require.Equal(t, app.Source, absPath)
}
