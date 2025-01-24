package generate

import (
	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/railwayapp/railpack-go/core/utils"
)

const (
	AptStepName = "apt"
)

type AptStepBuilder struct {
	DisplayName string
	DependsOn   []string
	Packages    []string
}

func (c *GenerateContext) NewAptStepBuilder(name string) *AptStepBuilder {
	step := &AptStepBuilder{
		DisplayName: c.GetStepName(name),
		DependsOn:   []string{},
		Packages:    []string{},
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *AptStepBuilder) AddAptPackage(pkg string) {
	b.Packages = append(b.Packages, pkg)
}

func (b *AptStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)
	step.DependsOn = utils.RemoveDuplicates(b.DependsOn)

	step.AddCommands([]plan.Command{
		options.NewAptInstallCommand(b.Packages),
	})

	return step, nil
}
