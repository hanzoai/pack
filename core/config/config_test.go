package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestMergeConfigSmall(t *testing.T) {
	config1JSON := `{
		"baseImage": "ubuntu:20.04",
		"packages": {
			"python": "latest",
			"node": "22"
		},
		"aptPackages": ["git"],
		"steps": {
			"install": {
				"dependsOn": ["packages"],
				"commands": [
					"echo first"
				]
			}
		}
	}`

	config2JSON := `{
		"baseImage": "secondd",
		"packages": {
			"node": "23",
			"bun": "latest"
		},
		"steps": {
			"install": {}
		}
	}`

	expectedJSON := `{
		"baseImage": "secondd",
		"packages": {
			"python": "latest",
			"node": "23",
			"bun": "latest"
		},
		"aptPackages": ["git"],
		"steps": {
			"install": {
				"dependsOn": ["packages"],
				"commands": [
					"echo first"
				]
			}
		},
		"caches": {}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	// fmt.Printf("CONFIG 1 COMMANDS: %+v\n", config1.Steps["install"].Commands)
	// fmt.Printf("CONFIG 2 COMMANDS: %+v\n", config2.Steps["install"].Commands)

	result := Merge(&config1, &config2)

	fmt.Printf("RESULT COMMANDS: %+v\n", result.Steps["install"].Commands)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}

}

func TestMergeConfig(t *testing.T) {
	config1JSON := `{
		"baseImage": "ubuntu:20.04",
		"packages": {
			"python": "latest",
			"node": "22"
		},
		"aptPackages": ["git"],
		"caches": {
			"npm": {
				"directory": "/root/.npm",
				"type": "locked"
			},
			"pip": {
				"directory": "/root/.cache/pip"
			}
		},
		"secrets": ["SECRET_1", "API_KEY"],
		"steps": {
			"install": {
				"name": "install",
				"useSecrets": true,
				"outputs": ["node_modules/", "package-lock.json"],
				"assets": {
					"package.json": "content1",
					"requirements.txt": "content2"
				},
				"startingImage": "node:16",
				"commands": [
					{"type": "exec", "cmd": "npm install", "caches": ["npm"], "customName": "Install NPM deps"},
					{"type": "path", "path": "/usr/local/bin"},
					{"type": "variable", "name": "NODE_ENV", "value": "production"},
					{"type": "copy", "src": "/src", "dest": "/app", "image": "alpine:latest"},
					{"type": "file", "path": "/app", "name": "config.json", "mode": 384, "customName": "Write config"}
				]
			},
			"build": {
				"name": "build",
				"commands": [
					{"type": "exec", "cmd": "config 1 a"},
					{"type": "exec", "cmd": "config 1 b"}
				]
			}
		},
		"start": {
			"command": "python app.py"
		}
	}`

	config2JSON := `{
		"providers": ["node"],
		"baseImage": "ubuntu:22.04",
		"packages": {
			"node": "23"
		},
		"aptPackages": ["curl"],
		"caches": {
			"npm": {
				"directory": "/root/.npm-new",
				"type": "shared"
			},
			"go": {
				"directory": "/root/.cache/go-build"
			}
		},
		"secrets": ["SECRET_2"],
		"steps": {
			"install": {
				"name": "install",
				"useSecrets": true,
				"outputs": ["dist/"],
				"assets": {
					"package.json": "content3"
				},
				"startingImage": "node:18"
			},
			"build": {
				"name": "build",
				"useSecrets": false,
				"commands": [
					{"type": "exec", "cmd": "config 2 a"},
					{"type": "exec", "cmd": "config 2 b"}
				]
			}
		},
		"start": {
			"baseImage": "node:18",
			"command": "node server.js",
			"paths": ["/usr/local/bin", "/app/bin"]
		}
	}`

	expectedJSON := `{
		"providers": ["node"],
		"baseImage": "ubuntu:22.04",
		"packages": {
			"python": "latest",
			"node": "23"
		},
		"aptPackages": ["curl"],
		"caches": {
			"npm": {
				"directory": "/root/.npm-new",
				"type": "shared"
			},
			"go": {
				"directory": "/root/.cache/go-build"
			},
			"pip": {
				"directory": "/root/.cache/pip"
			}
		},
		"secrets": ["SECRET_2"],
		"steps": {
			"install": {
				"name": "install",
				"useSecrets": true,
				"outputs": ["dist/"],
				"assets": {
					"package.json": "content3",
					"requirements.txt": "content2"
				},
				"startingImage": "node:18",
				"commands": [
					{"type": "exec", "cmd": "npm install", "caches": ["npm"], "customName": "Install NPM deps"},
					{"type": "path", "path": "/usr/local/bin"},
					{"type": "variable", "name": "NODE_ENV", "value": "production"},
					{"type": "copy", "src": "/src", "dest": "/app", "image": "alpine:latest"},
					{"type": "file", "path": "/app", "name": "config.json", "mode": 384, "customName": "Write config"}
				]
			},
			"build": {
				"name": "build",
				"useSecrets": false,
				"commands": [
					{"type": "exec", "cmd": "config 2 a"},
					{"type": "exec", "cmd": "config 2 b"}
				]
			}
		},
		"start": {
			"baseImage": "node:18",
			"command": "node server.js",
			"paths": ["/usr/local/bin", "/app/bin"]
		}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	fmt.Printf("CONFIG 1 COMMANDS: %+v\n", config1.Steps["install"].Commands)
	fmt.Printf("CONFIG 2 COMMANDS: %+v\n", config2.Steps["install"].Commands)

	result := Merge(&config1, &config2)

	fmt.Printf("RESULT COMMANDS: %+v\n", result.Steps["install"].Commands)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}
}

func TestMergeConfigStart(t *testing.T) {
	config1JSON := `{
		"start": {
			"command": "python app.py"
		}
	}`

	config2JSON := `{
		"packages": {
			"node": "23"
		}
	}`

	expectedJSON := `{
		"packages": {
			"node": "23"
		},
		"start": {
			"command": "python app.py"
		},
		"steps": {},
		"caches": {}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	result := Merge(&config1, &config2)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}
}

func TestGetJsonSchema(t *testing.T) {
	schema := GetJsonSchema()
	require.NotEmpty(t, schema)

	schemaJson, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)
	require.NotEmpty(t, schemaJson)
}
