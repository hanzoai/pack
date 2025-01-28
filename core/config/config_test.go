package config

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/plan"
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
		Steps: map[string]*plan.Step{
			"install": {
				Name: "install",
				Commands: []plan.Command{
					plan.NewExecCommand("npm install"),
				},
			},
			"build": {
				Name: "build",
				Commands: []plan.Command{
					plan.NewExecCommand("config 1 a"),
					plan.NewExecCommand("config 1 b"),
				},
			},
		},
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
		Steps: map[string]*plan.Step{
			"build": {
				Name: "build",
				Commands: []plan.Command{
					plan.NewExecCommand("config 2 a"),
					plan.NewExecCommand("config 2 b"),
				},
			},
		},
		Start: plan.Start{
			Command: "node server.js",
		},
	}

	result := config1.Merge(config2)

	require.Equal(t, "ubuntu:22.04", result.BaseImage)
	require.Equal(t, "latest", result.Packages["python"])
	require.Equal(t, "23", result.Packages["node"])
	require.ElementsMatch(t, []string{"git", "curl"}, result.AptPackages)
	require.Equal(t, "node server.js", result.Start.Command)
	require.Len(t, result.Steps["install"].Commands, 1)
	require.Equal(t, "npm install", result.Steps["install"].Commands[0].(plan.ExecCommand).Cmd)
	require.Len(t, result.Steps["build"].Commands, 2)
	require.Equal(t, "config 2 a", result.Steps["build"].Commands[0].(plan.ExecCommand).Cmd)
	require.Equal(t, "config 2 b", result.Steps["build"].Commands[1].(plan.ExecCommand).Cmd)

	config2.Start = plan.Start{}
	result = config1.Merge(config2)
	require.Equal(t, "python app.py", result.Start.Command)
}

func TestGetJsonSchema(t *testing.T) {
	schema := GetJsonSchema()
	require.NotEmpty(t, schema)

	schemaJson, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)
	require.NotEmpty(t, schemaJson)
}
