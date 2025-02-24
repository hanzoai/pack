package generate

import (
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

type ImageStepBuilder struct {
	DisplayName      string
	Resolver         *resolver.Resolver
	Packages         []*resolver.PackageRef
	ResolveStepImage func(options *BuildStepOptions) string
	AptPackages      []string
}

func (c *GenerateContext) NewImageStep(name string, resolveStepImage func(options *BuildStepOptions) string) *ImageStepBuilder {
	step := &ImageStepBuilder{
		DisplayName:      c.GetStepName(name),
		Resolver:         c.Resolver,
		ResolveStepImage: resolveStepImage,
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *ImageStepBuilder) Default(name string, defaultVersion string) resolver.PackageRef {
	for _, pkg := range b.Packages {
		if pkg.Name == name {
			return *pkg
		}
	}

	pkg := b.Resolver.Default(name, defaultVersion)
	b.Packages = append(b.Packages, &pkg)
	return pkg
}

func (b *ImageStepBuilder) Version(name resolver.PackageRef, version string, source string) {
	b.Resolver.Version(name, version, source)
}

func (b *ImageStepBuilder) Name() string {
	return b.DisplayName
}

func (b *ImageStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)

	image := b.ResolveStepImage(options)

	if len(b.AptPackages) > 0 {
		runtimeAptStep := plan.NewStep("packages:apt")
		runtimeAptStep.Inputs = []plan.Input{
			plan.NewImageInput(image),
		}
		runtimeAptStep.Commands = []plan.Command{
			options.NewAptInstallCommand(b.AptPackages),
		}

		step.Inputs = []plan.Input{
			plan.NewStepInput(runtimeAptStep.Name),
		}
	} else {
		step.Inputs = []plan.Input{
			plan.NewImageInput(image),
		}
	}

	return step, nil
}
