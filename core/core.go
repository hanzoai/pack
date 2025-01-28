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
	"github.com/railwayapp/railpack/core/utils"
)

const (
	defaultConfigFileName = "railpack.json"
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
	for _, provider := range providersToUse {
		err := provider.Plan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to run provider: %w", err)
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

	// Resolve all package versions into a fully qualified and valid version
	resolvedPackages, err := ctx.ResolvePackages()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve packages: %w", err)
	}

	// Generate the plan based on the context and resolved packages

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

// GetConfig merges the options, environment, and file config into a single config
func GetConfig(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) (*config.Config, error) {
	fmt.Println("Getting config")
	optionsConfig := GenerateConfigFromOptions(options)

	envConfig := GenerateConfigFromEnvironment(app, env)

	fileConfig, err := GenerateConfigFromFile(app, env)
	if err != nil {
		fmt.Println("No config file found")
		return nil, err
	}

	mergedConfig := config.Merge(optionsConfig, envConfig, fileConfig)

	return mergedConfig, nil
}

// GenerateConfigFromFile generates a config from the config file
func GenerateConfigFromFile(app *app.App, env *app.Environment) (*config.Config, error) {
	fmt.Println("Generating config from file")
	configFileName := defaultConfigFileName
	if envConfigFileName, _ := env.GetConfigVariable("CONFIG_FILE"); envConfigFileName != "" {
		configFileName = envConfigFileName
	}

	fmt.Println("Config file name:", configFileName)

	if !app.HasMatch(configFileName) {
		fmt.Printf("No config file found: %s\n", configFileName)
		log.Debugf("No config file found: %s", configFileName)
		return config.EmptyConfig(), nil
	}

	config := config.EmptyConfig()
	if err := app.ReadJSON(configFileName, config); err != nil {
		fmt.Printf("Failed to read config file: %s\n", err.Error())
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	fmt.Printf("Loaded config file: %s\n", configFileName)

	fmt.Printf("Config: %+v\n", config)

	fmt.Printf("Providers: %+v\n", config.Providers)

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

// GenerateConfigFromOptions generates a config from the CLI options
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
