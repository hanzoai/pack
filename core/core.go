package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack-go/core/app"
	"github.com/railwayapp/railpack-go/core/config"
	"github.com/railwayapp/railpack-go/core/generate"
	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/railwayapp/railpack-go/core/providers"
	"github.com/railwayapp/railpack-go/core/providers/procfile"
	"github.com/railwayapp/railpack-go/core/resolver"
	"github.com/railwayapp/railpack-go/core/utils"
)

type GenerateBuildPlanOptions struct {
	BuildCommand string
	StartCommand string
}

type BuildResult struct {
	Plan             *plan.BuildPlan                      `json:"plan"`
	ResolvedPackages map[string]*resolver.ResolvedPackage `json:"resolved_packages"`
	Metadata         map[string]string                    `json:"metadata"`
}

func GenerateBuildPlan(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) (*BuildResult, error) {
	ctx, err := generate.NewGenerateContext(app, env)
	if err != nil {
		return nil, err
	}

	config := GetConfig(app, env, options)

	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Debugf("Failed to serialize config: %v", err)
	} else {
		log.Debugf("Generated config:\n%s", string(configJson))
	}

	for _, provider := range providers.GetLanguageProviders() {
		matched, err := runProvider(provider, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to run provider: %w", err)
		}

		if matched {
			log.Debugf("Provider `%s` matched", provider.Name())
			ctx.Metadata.Set("provider", provider.Name())
			break
		}
	}

	procfileProvider := &procfile.ProcfileProvider{}
	if _, err := procfileProvider.Plan(ctx); err != nil {
		return nil, fmt.Errorf("failed to run procfile provider: %w", err)
	}

	if err := ApplyConfig(config, ctx); err != nil {
		return nil, fmt.Errorf("failed to apply config: %w", err)
	}

	resolvedPackages, err := ctx.ResolvePackages()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve packages: %w", err)
	}

	buildPlan := plan.NewBuildPlan()

	buildStepOptions := &generate.BuildStepOptions{
		ResolvedPackages: resolvedPackages,
		Caches:           ctx.Caches,
	}

	buildPlan.Variables = ctx.Variables
	for _, stepBuilder := range ctx.Steps {
		step, err := stepBuilder.Build(buildStepOptions)

		if err != nil {
			return nil, fmt.Errorf("failed to build step: %w", err)
		}

		buildPlan.AddStep(*step)
	}

	buildPlan.Caches = ctx.Caches.Caches

	buildPlan.Start.BaseImage = ctx.Start.BaseImage
	buildPlan.Start.Command = ctx.Start.Command
	buildPlan.Start.Paths = utils.RemoveDuplicates(ctx.Start.Paths)
	buildPlan.Start.Env = ctx.Start.Env

	buildResult := &BuildResult{
		Plan:             buildPlan,
		ResolvedPackages: resolvedPackages,
		Metadata:         ctx.Metadata.Properties,
	}

	return buildResult, nil
}

func runProvider(provider providers.Provider, ctx *generate.GenerateContext) (bool, error) {
	return provider.Plan(ctx)
}

func ApplyConfig(config *config.Config, ctx *generate.GenerateContext) error {
	log.Debugf("Applying config: %+v", config)

	// Mise package config
	miseStep := ctx.GetMiseStepBuilder()
	for pkg, version := range config.Packages {
		pkgRef := miseStep.Default(pkg, version)
		miseStep.Version(pkgRef, version, "config file")
	}

	// Apt package config
	if len(config.AptPackages) > 0 {
		aptStep := ctx.NewAptStepBuilder("config")
		aptStep.Packages = config.AptPackages
		miseStep.DependsOn = append(miseStep.DependsOn, aptStep.DisplayName)
	}

	// Step config
	for _, configStep := range config.Steps {
		if existingStep := ctx.GetStepByName(configStep.Name); existingStep != nil {
			existingStep.Commands = configStep.Commands
		} else {
			ctx.Steps = append(ctx.Steps, ctx.NewCommandStep(configStep.Name))
		}
	}

	// Cache config
	for name, cache := range config.Caches {
		ctx.Caches.SetCache(name, cache)
	}

	// Start config
	if config.Start.BaseImage != "" {
		ctx.Start.BaseImage = config.Start.BaseImage
	}

	if config.Start.Command != "" {
		ctx.Start.Command = config.Start.Command
	}

	if len(config.Start.Paths) > 0 {
		ctx.Start.Paths = append(ctx.Start.Paths, config.Start.Paths...)
	}

	if len(config.Start.Env) > 0 {
		if ctx.Start.Env == nil {
			ctx.Start.Env = make(map[string]string)
		}
		for k, v := range config.Start.Env {
			ctx.Start.Env[k] = v
		}
	}

	return nil
}

func GetConfig(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) *config.Config {
	optionsConfig := GenerateConfigFromOptions(options)
	envConfig := GenerateConfigFromEnvironment(app, env)

	mergedConfig := optionsConfig.Merge(envConfig)

	return mergedConfig
}

func GenerateConfigFromEnvironment(app *app.App, env *app.Environment) *config.Config {
	config := config.EmptyConfig()

	if env == nil {
		return config
	}

	if installCmdVar, _ := env.GetConfigVariable("INSTALL_CMD"); installCmdVar != "" {
		installStep := config.GetOrCreateStep("install")
		installStep.Commands = []plan.Command{plan.NewExecCommand(installCmdVar)}
		installStep.DependsOn = append(installStep.DependsOn, "packages")
	}

	if buildCmdVar, _ := env.GetConfigVariable("BUILD_CMD"); buildCmdVar != "" {
		buildStep := config.GetOrCreateStep("build")
		buildStep.Commands = []plan.Command{plan.NewExecCommand(buildCmdVar)}
		buildStep.DependsOn = append(buildStep.DependsOn, "install")
	}

	if startCmdVar, _ := env.GetConfigVariable("START_CMD"); startCmdVar != "" {
		config.Start.Command = startCmdVar
	}

	if envPackages, _ := env.GetConfigVariable("PACKAGES"); envPackages != "" {
		config.Packages = make(map[string]string)
		for _, pkg := range strings.Split(envPackages, " ") {
			config.Packages[pkg] = "latest"
		}
	}

	if envAptPackages, _ := env.GetConfigVariable("APT_PACKAGES"); envAptPackages != "" {
		config.AptPackages = strings.Split(envAptPackages, " ")
	}

	return config
}

func GenerateConfigFromOptions(options *GenerateBuildPlanOptions) *config.Config {
	config := config.EmptyConfig()

	if options == nil {
		return config
	}

	if options.BuildCommand != "" {
		buildStep := config.GetOrCreateStep("build")
		buildStep.Commands = []plan.Command{plan.NewExecCommand(options.BuildCommand)}
		buildStep.DependsOn = append(buildStep.DependsOn, "install")
	}

	if options.StartCommand != "" {
		config.Start.Command = options.StartCommand
	}

	return config
}
