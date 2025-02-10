package core

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers"
	"github.com/railwayapp/railpack/core/providers/procfile"
	"github.com/railwayapp/railpack/core/resolver"
)

const (
	defaultConfigFileName = "railpack.json"
)

type GenerateBuildPlanOptions struct {
	BuildCommand     string
	StartCommand     string
	PreviousVersions map[string]string
}

type BuildResult struct {
	Plan             *plan.BuildPlan                      `json:"plan,omitempty"`
	ResolvedPackages map[string]*resolver.ResolvedPackage `json:"resolvedPackages,omitempty"`
	Metadata         map[string]string                    `json:"metadata,omitempty"`
}

func GenerateBuildPlan(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) (*BuildResult, error) {
	ctx, err := generate.NewGenerateContext(app, env)
	if err != nil {
		return nil, err
	}

	// Set the preivous versions
	if options.PreviousVersions != nil {
		for name, version := range options.PreviousVersions {
			ctx.Resolver.SetPreviousVersion(name, version)
		}
	}

	// Get the full user config based on file config, env config, and options
	config, err := GetConfig(app, env, options)
	if err != nil {
		return nil, err
	}

	// Figure out what providers to use
	providersToUse := getProviders(ctx, config)
	providerNames := make([]string, len(providersToUse))
	for i, provider := range providersToUse {
		providerNames[i] = provider.Name()
	}
	ctx.Metadata.Set("providers", strings.Join(providerNames, ","))

	// Run the providers to update the context with how to build the app
	for i, provider := range providersToUse {
		// If this is not the first provider, we need to enter a subcontext so that step names are unique
		if i > 0 {
			ctx.EnterSubContext(provider.Name())
		}

		err := provider.Plan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to run provider: %w", err)
		}

		if i > 0 {
			ctx.ExitSubContext()
		}
	}

	// Run the procfile provider to support apps that have a Procfile with a start command
	procfileProvider := &procfile.ProcfileProvider{}
	if _, err := procfileProvider.Plan(ctx); err != nil {
		return nil, fmt.Errorf("failed to run procfile provider: %w", err)
	}

	// Update the context with the config
	if err := ctx.ApplyConfig(config); err != nil {
		return nil, fmt.Errorf("failed to apply config: %w", err)
	}

	buildPlan, resolvedPackages, err := ctx.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate build plan: %w", err)
	}

	buildResult := &BuildResult{
		Plan:             buildPlan,
		ResolvedPackages: resolvedPackages,
		Metadata:         ctx.Metadata.Properties,
	}

	return buildResult, nil
}

// GetConfig merges the options, environment, and file config into a single config
func GetConfig(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) (*config.Config, error) {
	optionsConfig := GenerateConfigFromOptions(options)

	envConfig := GenerateConfigFromEnvironment(app, env)

	fileConfig, err := GenerateConfigFromFile(app, env)
	if err != nil {
		return nil, err
	}

	mergedConfig := config.Merge(optionsConfig, envConfig, fileConfig)

	return mergedConfig, nil
}

// GenerateConfigFromFile generates a config from the config file
func GenerateConfigFromFile(app *app.App, env *app.Environment) (*config.Config, error) {
	configFileName := defaultConfigFileName
	if envConfigFileName, _ := env.GetConfigVariable("CONFIG_FILE"); envConfigFileName != "" {
		configFileName = envConfigFileName
	}

	if !app.HasMatch(configFileName) {
		return config.EmptyConfig(), nil
	}

	config := config.EmptyConfig()
	if err := app.ReadJSON(configFileName, config); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return config, nil
}

// GenerateConfigFromEnvironment generates a config from the environment
func GenerateConfigFromEnvironment(app *app.App, env *app.Environment) *config.Config {
	config := config.EmptyConfig()

	if env == nil {
		return config
	}

	if installCmdVar, _ := env.GetConfigVariable("INSTALL_CMD"); installCmdVar != "" {
		installStep := config.GetOrCreateStep("install")
		installStep.Commands = &[]plan.Command{plan.NewExecCommand(installCmdVar)}
		installStep.DependsOn = append(installStep.DependsOn, "packages")
	}

	if buildCmdVar, _ := env.GetConfigVariable("BUILD_CMD"); buildCmdVar != "" {
		buildStep := config.GetOrCreateStep("build")
		buildStep.Commands = &[]plan.Command{plan.NewExecCommand(buildCmdVar)}
		buildStep.DependsOn = append(buildStep.DependsOn, "install")
	}

	if startCmdVar, _ := env.GetConfigVariable("START_CMD"); startCmdVar != "" {
		config.Start.Command = startCmdVar
	}

	if envPackages, _ := env.GetConfigVariable("PACKAGES"); envPackages != "" {
		config.Packages = make(map[string]string)
		for _, pkg := range strings.Split(envPackages, " ") {
			// TODO: We should support specifying a version here (e.g. "node@18" or just "node")
			config.Packages[pkg] = "latest"
		}
	}

	if envAptPackages, _ := env.GetConfigVariable("APT_PACKAGES"); envAptPackages != "" {
		config.AptPackages = strings.Split(envAptPackages, " ")
	}

	for name := range env.Variables {
		config.Secrets = append(config.Secrets, name)
	}

	return config
}

// GenerateConfigFromOptions generates a config from the CLI options
func GenerateConfigFromOptions(options *GenerateBuildPlanOptions) *config.Config {
	config := config.EmptyConfig()

	if options == nil {
		return config
	}

	if options.BuildCommand != "" {
		buildStep := config.GetOrCreateStep("build")
		buildStep.Commands = &[]plan.Command{plan.NewExecCommand(options.BuildCommand)}
		buildStep.DependsOn = append(buildStep.DependsOn, "install")
	}

	if options.StartCommand != "" {
		config.Start.Command = options.StartCommand
	}

	return config
}

func getProviders(ctx *generate.GenerateContext, config *config.Config) []providers.Provider {
	var providersToUse []providers.Provider

	allProviders := providers.GetLanguageProviders()

	// If there are no providers manually specified in the config,
	// use the first provider that is detected
	if config.Providers == nil {
		for _, provider := range allProviders {
			matched, err := provider.Detect(ctx)
			if err != nil {
				log.Warnf("Failed to detect provider `%s`: %s", provider.Name(), err.Error())
				continue
			}

			if matched {
				providersToUse = append(providersToUse, provider)
				break
			}
		}

		return providersToUse
	}

	// Otherwise, use the providers specified in the config
	for _, providerName := range *config.Providers {
		provider := providers.GetProvider(providerName)
		if provider == nil {
			log.Warnf("Provider `%s` not found", providerName)
			continue
		}

		providersToUse = append(providersToUse, provider)
	}

	return providersToUse
}
