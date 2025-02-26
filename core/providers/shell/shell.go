package shell

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	StartScriptName = "start.sh"
)

type ShellProvider struct{}

func (p *ShellProvider) Name() string {
	return "shell"
}

func (p *ShellProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	// Check if start.sh exists at the root
	return ctx.App.HasMatch(StartScriptName), nil
}

func (p *ShellProvider) Initialize(ctx *generate.GenerateContext) error {
	// No initialization needed
	return nil
}

func (p *ShellProvider) Plan(ctx *generate.GenerateContext) error {
	// Set the start command to run the start.sh script
	ctx.Deploy.StartCmd = "sh " + StartScriptName

	// Add metadata
	ctx.Metadata.Set("shellScript", StartScriptName)

	// Make sure the script is executable
	setup := ctx.NewCommandStep("setup")
	setup.AddInput(ctx.DefaultRuntimeInput())
	setup.AddCommands(
		[]plan.Command{
			plan.NewCopyCommand(StartScriptName),
			plan.NewExecCommand("chmod +x " + StartScriptName),
			plan.NewExecCommand("sh " + StartScriptName),
		},
	)

	// Add the script to the deploy inputs
	ctx.Deploy.Inputs = []plan.Input{
		plan.NewStepInput(setup.Name()),
		plan.NewStepInput(setup.Name(), plan.InputOptions{
			Include: []string{"."},
		}),
		plan.NewLocalInput("."),
	}

	return nil
}
