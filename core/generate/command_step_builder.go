package generate

import (
	"maps"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/utils"
)

type CommandStepBuilder struct {
	DisplayName string
	DependsOn   []string
	Commands    *[]plan.Command
	Outputs     *[]string
	Assets      map[string]string
	Variables   map[string]string
	Caches      []string
	UseSecrets  bool
}

func (c *GenerateContext) NewCommandStep(name string) *CommandStepBuilder {
	step := &CommandStepBuilder{
		DisplayName: c.GetStepName(name),
		DependsOn:   []string{MisePackageStepName},
		Commands:    &[]plan.Command{},
		Assets:      map[string]string{},
		Variables:   map[string]string{},
		Caches:      []string{},
		UseSecrets:  true,
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *CommandStepBuilder) DependOn(name string) {
	b.DependsOn = append(b.DependsOn, name)
}

func (b *CommandStepBuilder) AddVariables(variables map[string]string) {
	maps.Copy(b.Variables, variables)
}

func (b *CommandStepBuilder) AddCache(name string) {
	b.Caches = append(b.Caches, name)
}

func (b *CommandStepBuilder) AddCommand(command plan.Command) {
	if b.Commands == nil {
		b.Commands = &[]plan.Command{}
	}
	*b.Commands = append(*b.Commands, command)
}

func (b *CommandStepBuilder) AddCommands(commands []plan.Command) {
	if b.Commands == nil {
		b.Commands = &[]plan.Command{}
	}
	*b.Commands = append(*b.Commands, commands...)
}

func (b *CommandStepBuilder) AddEnvVars(envVars map[string]string) {
	maps.Copy(b.Variables, envVars)
}

func (b *CommandStepBuilder) AddPaths(paths []string) {
	commands := []plan.Command{}
	for _, path := range paths {
		commands = append(commands, plan.NewPathCommand(path))
	}
	b.AddCommands(commands)
}

func (b *CommandStepBuilder) Name() string {
	return b.DisplayName
}

func (b *CommandStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)

	step.DependsOn = utils.RemoveDuplicates(b.DependsOn)
	step.Outputs = b.Outputs
	step.Commands = b.Commands
	step.Assets = b.Assets
	step.Caches = b.Caches
	step.Variables = b.Variables

	if !b.UseSecrets {
		step.UseSecrets = &b.UseSecrets
	}

	return step, nil
}
