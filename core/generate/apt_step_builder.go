package generate

import (
	"fmt"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/utils"
)

type AptStepBuilder struct {
	DisplayName string
	DependsOn   []string
	Packages    []string
}

func (c *GenerateContext) NewAptStepBuilder(name string) *AptStepBuilder {
	step := &AptStepBuilder{
		DisplayName: c.GetStepName(fmt.Sprintf("packages:apt:%s", name)),
		DependsOn:   []string{},
		Packages:    []string{},
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *AptStepBuilder) AddAptPackage(pkg string) {
	b.Packages = append(b.Packages, pkg)
}

func (b *AptStepBuilder) Name() string {
	return b.DisplayName
}

func (b *AptStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)
	step.DependsOn = utils.RemoveDuplicates(b.DependsOn)

	step.AddCommands([]plan.Command{
		options.NewAptInstallCommand(b.Packages),
	})

	step.Caches = options.Caches.GetAptCaches()

	// Does not use any secrets
	step.Secrets = &[]string{}

	return step, nil
}
