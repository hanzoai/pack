package generate

import (
	"strings"

	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

type BuildStepOptions struct {
	ResolvedPackages map[string]*resolver.ResolvedPackage
	Caches           *CacheContext
}

type StepBuilder interface {
	Name() string
	Build(options *BuildStepOptions) (*plan.Step, error)
}

type GenerateContext struct {
	App *a.App
	Env *a.Environment

	Steps []StepBuilder
	Start StartContext

	Caches    *CacheContext
	Variables map[string]string

	SubContexts []string

	Metadata *Metadata

	resolver        *resolver.Resolver
	miseStepBuilder *MiseStepBuilder
}

func NewGenerateContext(app *a.App, env *a.Environment) (*GenerateContext, error) {
	resolver, err := resolver.NewResolver(mise.TestInstallDir)
	if err != nil {
		return nil, err
	}

	return &GenerateContext{
		App:       app,
		Env:       env,
		Variables: map[string]string{},
		Steps:     make([]StepBuilder, 0),
		Start:     *NewStartContext(),
		Caches:    NewCacheContext(),
		Metadata:  NewMetadata(),
		resolver:  resolver,
	}, nil
}

func (c *GenerateContext) GetMiseStepBuilder() *MiseStepBuilder {
	if c.miseStepBuilder == nil {
		c.miseStepBuilder = c.newMiseStepBuilder()
	}
	return c.miseStepBuilder
}

func (c *GenerateContext) EnterSubContext(subContext string) *GenerateContext {
	c.SubContexts = append(c.SubContexts, subContext)
	return c
}

func (c *GenerateContext) ExitSubContext() *GenerateContext {
	c.SubContexts = c.SubContexts[:len(c.SubContexts)-1]
	return c
}

func (c *GenerateContext) GetStepName(name string) string {
	subContextNames := strings.Join(c.SubContexts, ":")
	if subContextNames != "" {
		return name + ":" + subContextNames
	}
	return name
}

func (c *GenerateContext) GetStepByName(name string) *StepBuilder {
	for _, step := range c.Steps {
		if step.Name() == name {
			return &step
		}
	}
	return nil
}

func (c *GenerateContext) ResolvePackages() (map[string]*resolver.ResolvedPackage, error) {
	return c.resolver.ResolvePackages()
}

func (o *BuildStepOptions) NewAptInstallCommand(pkgs []string) plan.Command {
	return plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+strings.Join(pkgs, " ")+"'", plan.ExecOptions{
		CustomName: "install apt packages: " + strings.Join(pkgs, " "),
		Caches:     o.Caches.GetAptCaches(),
	})
}
