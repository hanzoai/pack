package generate

import (
	"encoding/json"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
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
	ctx := CreateTestContext(t, "../../examples/node-npm")
	provider := &TestProvider{}
	require.NoError(t, provider.Plan(ctx))

	// User defined config
	configJSON := `{
		"packages": {
			"node": "20.18.2",
			"go": "1.23.5",
			"python": "3.13.1"
		},
		"aptPackages": ["curl"],
		"steps": {
			"build": {
				"commands": ["echo building"]
			}
		},
		"secrets": ["RAILWAY_SECRET_1", "RAILWAY_SECRET_2"],
		"start": {
			"variables": {
				"HELLO": "world"
			}
		}
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
	for _, builder := range ctx.Steps {
		_, err := builder.Build(buildOpts)
		require.NoError(t, err)
	}

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)

	buildPlanJSON, err := json.Marshal(buildPlan)
	require.NoError(t, err)

	var actualPlan map[string]interface{}
	require.NoError(t, json.Unmarshal(buildPlanJSON, &actualPlan))

	serializedPlan, err := json.MarshalIndent(actualPlan, "", "  ")
	require.NoError(t, err)

	snaps.MatchJSON(t, serializedPlan)
}
