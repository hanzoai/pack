package generate

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
	"github.com/railwayapp/railpack/core/utils"
)

const (
	DefaultBaseImage = "ghcr.io/railwayapp/railpack-runtime-base:latest"
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

	BaseImage string
	Steps     []StepBuilder
	Start     StartContext

	Caches  *CacheContext
	Secrets []string

	SubContexts []string

	Metadata *Metadata

	Resolver *resolver.Resolver

	miseStepBuilder *MiseStepBuilder
}

func NewGenerateContext(app *a.App, env *a.Environment) (*GenerateContext, error) {
	resolver, err := resolver.NewResolver(mise.InstallDir)
	if err != nil {
		return nil, err
	}

	return &GenerateContext{
		App:       app,
		Env:       env,
		BaseImage: DefaultBaseImage,
		Steps:     make([]StepBuilder, 0),
		Start:     *NewStartContext(),
		Caches:    NewCacheContext(),
		Secrets:   []string{},
		Metadata:  NewMetadata(),
		Resolver:  resolver,
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
	return c.Resolver.ResolvePackages()
}

// Generate a build plan from the context
func (c *GenerateContext) Generate() (*plan.BuildPlan, map[string]*resolver.ResolvedPackage, error) {
	// Resolve all package versions into a fully qualified and valid version
	resolvedPackages, err := c.ResolvePackages()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve packages: %w", err)
	}

	// Generate the plan based on the context and resolved packages

	buildPlan := plan.NewBuildPlan()
	buildPlan.BaseImage = c.BaseImage

	buildStepOptions := &BuildStepOptions{
		ResolvedPackages: resolvedPackages,
		Caches:           c.Caches,
	}

	for _, stepBuilder := range c.Steps {
		step, err := stepBuilder.Build(buildStepOptions)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to build step: %w", err)
		}

		buildPlan.AddStep(*step)
	}

	buildPlan.Caches = c.Caches.Caches

	buildPlan.Secrets = utils.RemoveDuplicates(c.Secrets)

	buildPlan.Start.BaseImage = c.Start.BaseImage
	buildPlan.Start.Command = c.Start.Command
	buildPlan.Start.Outputs = utils.RemoveDuplicates(c.Start.outputs)
	buildPlan.Start.Paths = utils.RemoveDuplicates(c.Start.paths)
	buildPlan.Start.Variables = c.Start.variables

	return buildPlan, resolvedPackages, nil
}

func (c *GenerateContext) ApplyConfig(config *config.Config) error {
	// Mise package config
	miseStep := c.GetMiseStepBuilder()
	for _, pkg := range slices.Sorted(maps.Keys(config.Packages)) {
		version := config.Packages[pkg]
		pkgRef := miseStep.Default(pkg, version)
		miseStep.Version(pkgRef, version, "custom config")
	}

	// Apt package config
	aptStepName := ""
	if len(config.AptPackages) > 0 {
		aptStep := c.NewAptStepBuilder("config")
		aptStepName = aptStep.Name()
		aptStep.Packages = config.AptPackages

		// We install the apt packages again in the mise step since they may be required for install mise packages
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, config.AptPackages...)
	}

	// Step config
	for _, name := range slices.Sorted(maps.Keys(config.Steps)) {
		configStep := config.Steps[name]

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

		if aptStepName != "" {
			commandStepBuilder.DependsOn = append(commandStepBuilder.DependsOn, aptStepName)
		}

		if configStep.Commands != nil {
			commandStepBuilder.Commands = configStep.Commands
		}

		if configStep.Outputs != nil {
			commandStepBuilder.Outputs = configStep.Outputs
		}
		for k, v := range configStep.Assets {
			commandStepBuilder.Assets[k] = v
		}

		if configStep.UseSecrets != nil {
			commandStepBuilder.UseSecrets = *configStep.UseSecrets
		}

		if len(configStep.Caches) > 0 {
			commandStepBuilder.Caches = configStep.Caches
		}

		if configStep.Variables != nil {
			commandStepBuilder.AddEnvVars(configStep.Variables)
		}
	}

	// Cache config
	for name, cache := range config.Caches {
		c.Caches.SetCache(name, cache)
	}

	// Secret config
	c.Secrets = append(c.Secrets, config.Secrets...)

	// Start config
	if config.Start.BaseImage != "" {
		c.Start.BaseImage = config.Start.BaseImage
	}

	if config.Start.Command != "" {
		c.Start.Command = config.Start.Command
	}

	if config.Start.Variables != nil {
		c.Start.AddEnvVars(config.Start.Variables)
	}

	c.Start.AddPaths(config.Start.Paths)

	if len(config.Start.Outputs) > 0 {
		c.Start.outputs = config.Start.Outputs
	}

	return nil
}

func (o *BuildStepOptions) NewAptInstallCommand(pkgs []string) plan.Command {
	pkgs = utils.RemoveDuplicates(pkgs)
	sort.Strings(pkgs)

	return plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+strings.Join(pkgs, " ")+"'", plan.ExecOptions{
		CustomName: "install apt packages: " + strings.Join(pkgs, " "),
	})
}
