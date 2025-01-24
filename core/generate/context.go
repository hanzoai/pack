package generate

import (
	"strings"

	a "github.com/railwayapp/railpack-go/core/app"
	"github.com/railwayapp/railpack-go/core/mise"
	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/railwayapp/railpack-go/core/resolver"
)

type BuildStepOptions struct {
	ResolvedPackages map[string]*resolver.ResolvedPackage
	Caches           map[string]*plan.Cache
}

type StepBuilder interface {
	Build(options *BuildStepOptions) (*plan.Step, error)
}

type GenerateContext struct {
	App *a.App
	Env *a.Environment

	Steps []StepBuilder
	Start StartContext

	Caches    map[string]*plan.Cache
	Variables map[string]string

	SubContexts []string

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
		Caches:    make(map[string]*plan.Cache),
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

func (c *GenerateContext) ResolvePackages() (map[string]*resolver.ResolvedPackage, error) {
	return c.resolver.ResolvePackages()
}

func (c *GenerateContext) AddCache(key string, directory string) string {
	c.Caches[key] = plan.NewCache(directory)
	return key
}
