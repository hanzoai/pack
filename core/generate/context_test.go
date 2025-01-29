package generate

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

type TestProvider struct{}

func (p *TestProvider) Plan(ctx *GenerateContext) error {
	// mise
	mise := ctx.GetMiseStepBuilder()
	nodeRef := mise.Default("node", "18")
	mise.Version(nodeRef, "18", "test")

	// apt
	aptStep := ctx.NewAptStepBuilder("test")
	aptStep.AddAptPackage("git")
	aptStep.AddAptPackage("neofetch")

	// commands
	installStep := ctx.NewCommandStep("install")
	installStep.AddCommand(plan.NewExecCommand("npm install", plan.ExecOptions{}))
	installStep.Outputs = &[]string{"node_modules"}
	installStep.DependsOn = []string{aptStep.Name()}

	buildStep := ctx.NewCommandStep("build")
	buildStep.AddCommand(plan.NewExecCommand("npm run build", plan.ExecOptions{}))
	buildStep.DependsOn = []string{installStep.Name()}

	ctx.Start.Command = "npm run start"

	return nil
}

func CreateTestContext(t *testing.T, path string) *GenerateContext {
	t.Helper()

	userApp, err := app.NewApp(path)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)

	ctx, err := NewGenerateContext(userApp, env)
	require.NoError(t, err)

	return ctx
}

func TestGenerateContext(t *testing.T) {
	ctx := CreateTestContext(t, "../../examples/node-npm-latest")
	provider := &TestProvider{}
	require.NoError(t, provider.Plan(ctx))

	// User defined config
	configJSON := `{
		"packages": {
			"node": "20",
			"python": "3.11"
		},
		"aptPackages": ["curl"],
		"steps": {
			"build": {
				"commands": ["echo building"]
			}
		},
		"secrets": ["RAILWAY_SECRET_1", "RAILWAY_SECRET_2"]
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))

	// Apply the config to the context
	require.NoError(t, ctx.ApplyConfig(&cfg))

	// Resolve packages
	resolvedPkgs, err := ctx.ResolvePackages()
	require.NoError(t, err)

	// Generate a plan
	buildOpts := &BuildStepOptions{
		ResolvedPackages: resolvedPkgs,
		Caches:           ctx.Caches,
	}

	// Build and verify each step
	var steps []map[string]interface{}
	for _, builder := range ctx.Steps {
		step, err := builder.Build(buildOpts)
		require.NoError(t, err)

		stepJSON, err := json.Marshal(step)
		require.NoError(t, err)

		var stepMap map[string]interface{}
		require.NoError(t, json.Unmarshal(stepJSON, &stepMap))
		steps = append(steps, stepMap)
	}

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)

	buildPlanJSON, err := json.Marshal(buildPlan)
	require.NoError(t, err)

	var actualPlan map[string]interface{}
	require.NoError(t, json.Unmarshal(buildPlanJSON, &actualPlan))

	_, err = json.MarshalIndent(actualPlan, "", "  ")
	require.NoError(t, err)

	// Expected steps after build
	expectedPlanJSON := `{
  "steps": [
    {
      "assets": {
        "mise.toml": "[tools]\n  [tools.node]\n    version = \"20.18.2\"\n  [tools.python]\n    version = \"3.11.11\"\n"
      },
      "commands": [
        { "name": "MISE_DATA_DIR", "value": "/mise" },
        { "name": "MISE_CONFIG_DIR", "value": "/mise" },
        { "name": "MISE_INSTALL_PATH", "value": "/usr/local/bin/mise" },
        { "name": "MISE_CACHE_DIR", "value": "/mise/cache" },
        { "path": "/mise/shims" },
        { "caches": ["apt", "apt-lists"], "cmd": "sh -c 'apt-get update \u0026\u0026 apt-get install -y curl ca-certificates'", "customName": "install apt packages: curl ca-certificates" },
        { "caches": ["mise"], "cmd": "sh -c 'curl -fsSL https://mise.run | sh'", "customName": "install mise" },
        { "customName": "create mise config", "name": "mise.toml", "path": "/etc/mise/config.toml" },
        { "caches": ["mise"], "cmd": "sh -c 'mise trust -a \u0026\u0026 mise install'", "customName": "install mise packages: node, python" }
      ],
      "dependsOn": ["packages:apt:config"],
      "name": "packages:mise",
      "outputs": ["/mise/shims", "/mise/installs", "/usr/local/bin/mise", "/etc/mise/config.toml", "/root/.local/state/mise"]
    },
    {
      "commands": [
        { "caches": ["apt", "apt-lists"], "cmd": "sh -c 'apt-get update \u0026\u0026 apt-get install -y git neofetch'", "customName": "install apt packages: git neofetch" }
      ],
      "name": "packages:apt:test"
    },
    {
      "commands": [{ "cmd": "npm install" }],
      "dependsOn": ["packages:apt:test"],
      "name": "install",
      "outputs": ["node_modules"]
    },
    {
      "commands": [{ "cmd": "npm run build" }, { "cmd": "echo building" }],
      "dependsOn": ["install"],
      "name": "build"
    },
    {
      "commands": [
        { "caches": ["apt", "apt-lists"], "cmd": "sh -c 'apt-get update \u0026\u0026 apt-get install -y curl'", "customName": "install apt packages: curl" }
      ],
      "name": "packages:apt:config"
    }
  ],
  "start": {
    "cmd": "npm run start"
  },
  "secrets": ["RAILWAY_SECRET_1", "RAILWAY_SECRET_2"],
  "caches": {
    "apt": { "directory": "/var/cache/apt", "type": "locked" },
    "apt-lists": { "directory": "/var/lib/apt/lists", "type": "locked" },
    "mise": { "directory": "/mise/cache", "type": "shared" }
  }
}`

	var expectedPlan map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(expectedPlanJSON), &expectedPlan))

	if diff := cmp.Diff(expectedPlan, actualPlan); diff != "" {
		t.Errorf("plan mismatch (-want +got):\n%s", diff)
	}
}
