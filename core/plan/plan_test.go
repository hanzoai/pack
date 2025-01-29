package plan

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestSerialization(t *testing.T) {
	jsonPlan := `{
		"steps": [
			{
				"name": "install",
				"commands": [
					{"cmd": "apt-get update"},
					{"cmd": "apt-get install -y curl"}
				],
				"startingImage": "ubuntu:22.04"
			},
			{
				"name": "deps",
				"dependsOn": ["install"],
				"commands": [
					{"path": "/root/.npm", "name": ".npmrc"},
					{"cmd": "npm ci"},
					{"cmd": "npm run build"}
				],
				"useSecrets": true,
				"outputs": [
					"dist",
					"node_modules/.cache"
				],
				"assets": {
					"npmrc": "registry=https://registry.npmjs.org/\n//registry.npmjs.org/:_authToken=${NPM_TOKEN}\nalways-auth=true"
				}
			},
			{
				"name": "build",
				"dependsOn": ["deps"],
				"commands": [
					{"src": ".", "dest": "."},
					{"cmd": "npm run test"},
					{"path": "/usr/local/bin"},
					{"name": "NODE_ENV", "value": "production"}
				],
				"useSecrets": false
			}
		],
		"start": {
			"baseImage": "node:18-slim",
			"cmd": "npm start",
			"paths": ["/usr/local/bin", "/app/node_modules/.bin"]
		},
		"caches": {
			"npm": {
				"directory": "/root/.npm",
				"type": "shared"
			},
			"build-cache": {
				"directory": "node_modules/.cache",
				"type": "locked"
			}
		},
		"secrets": ["NPM_TOKEN", "GITHUB_TOKEN"]
	}`

	var plan1 BuildPlan
	err := json.Unmarshal([]byte(jsonPlan), &plan1)
	require.NoError(t, err)

	serialized, err := json.MarshalIndent(&plan1, "", "  ")
	require.NoError(t, err)

	var plan2 BuildPlan
	err = json.Unmarshal(serialized, &plan2)
	require.NoError(t, err)

	if diff := cmp.Diff(plan1, plan2); diff != "" {
		t.Errorf("plans mismatch (-want +got):\n%s", diff)
	}
}
