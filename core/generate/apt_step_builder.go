package generate

import (
	"strings"

	"github.com/railwayapp/railpack-go/core/plan"
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
	step.DependsOn = b.DependsOn

	pkgString := strings.Join(b.Packages, " ")
	step.AddCommands([]plan.Command{
		plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+pkgString+" && rm -rf /var/lib/apt/lists/*'",
			plan.ExecOptions{CustomName: "install apt packages: " + pkgString}),
	})

	return step, nil
}
