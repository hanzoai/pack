package generate

import (
	"strings"

	"github.com/charmbracelet/log"
	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
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

func (c *GenerateContext) ApplyConfig(config *config.Config) error {
	// Mise package config
	miseStep := c.GetMiseStepBuilder()
	for pkg, version := range config.Packages {
		pkgRef := miseStep.Default(pkg, version)
		miseStep.Version(pkgRef, version, "custom config")
	}

	// Apt package config
	if len(config.AptPackages) > 0 {
		aptStep := c.NewAptStepBuilder("config")
		aptStep.Packages = config.AptPackages

		// The apt step should run first
		miseStep.DependsOn = append(miseStep.DependsOn, aptStep.DisplayName)
	}

	// Step config
	for name, configStep := range config.Steps {
		var commandStepBuilder *CommandStepBuilder

		// We need to use the key as the step name and not `configStep.Name`
		if existingStep := c.GetStepByName(name); existingStep != nil {
			if csb, ok := (*existingStep).(*CommandStepBuilder); ok {
				commandStepBuilder = csb
			} else {
				log.Warnf("Step `%s` exists, but it is not a command step. Skipping...", name)
				continue
			}
		} else {
			commandStepBuilder = c.NewCommandStep(name)
		}

		// Overwrite the step with values from the config if they exist
		if len(configStep.DependsOn) > 0 {
			commandStepBuilder.DependsOn = configStep.DependsOn
		}
		if len(configStep.Commands) > 0 {
			commandStepBuilder.Commands = configStep.Commands
		}
		if len(configStep.Outputs) > 0 {
			commandStepBuilder.Outputs = configStep.Outputs
		}
		for k, v := range configStep.Assets {
			commandStepBuilder.Assets[k] = v
		}
	}

	// Cache config
	for name, cache := range config.Caches {
		c.Caches.SetCache(name, cache)
	}

	// Start config
	if config.Start.BaseImage != "" {
		c.Start.BaseImage = config.Start.BaseImage
	}

	if config.Start.Command != "" {
		c.Start.Command = config.Start.Command
	}

	if len(config.Start.Paths) > 0 {
		c.Start.Paths = append(c.Start.Paths, config.Start.Paths...)
	}

	if len(config.Start.Env) > 0 {
		if c.Start.Env == nil {
			c.Start.Env = make(map[string]string)
		}
		for k, v := range config.Start.Env {
			c.Start.Env[k] = v
		}
	}

	return nil
}

func (o *BuildStepOptions) NewAptInstallCommand(pkgs []string) plan.Command {
	return plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+strings.Join(pkgs, " ")+"'", plan.ExecOptions{
		CustomName: "install apt packages: " + strings.Join(pkgs, " "),
		Caches:     o.Caches.GetAptCaches(),
	})
}
