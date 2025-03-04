package core

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
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
	RailpackVersion          string
	BuildCommand             string
	StartCommand             string
	PreviousVersions         map[string]string
	ConfigFilePath           string
	ErrorMissingStartCommand bool
}

type BuildResult struct {
	RailpackVersion   string                               `json:"railpackVersion,omitempty"`
	Plan              *plan.BuildPlan                      `json:"plan,omitempty"`
	ResolvedPackages  map[string]*resolver.ResolvedPackage `json:"resolvedPackages,omitempty"`
	Metadata          map[string]string                    `json:"metadata,omitempty"`
	DetectedProviders []string                             `json:"detectedProviders,omitempty"`
	Logs              []logger.Msg                         `json:"logs,omitempty"`
	Success           bool                                 `json:"success,omitempty"`
}

func GenerateBuildPlan(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions) *BuildResult {
	logger := logger.NewLogger()

	// Get the full user config based on file config, env config, and options
	config, err := GetConfig(app, env, options, logger)
	if err != nil {
		logger.LogError("%s", err.Error())
		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	ctx, err := generate.NewGenerateContext(app, env, config, logger)
	if err != nil {
		logger.LogError("%s", err.Error())
		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	// Set the preivous versions
	if options.PreviousVersions != nil {
		for name, version := range options.PreviousVersions {
			ctx.Resolver.SetPreviousVersion(name, version)
		}
	}

	// Figure out what providers to use
	providerToUse, detectedProviderName := getProviders(ctx, config)
	ctx.Metadata.Set("providers", detectedProviderName)

	// TODO: We should indicate if we have packages specified in the config
	// so that providers can determine if they should include mise in the final image (e.g. for shell script)

	if providerToUse != nil {
		err = providerToUse.Plan(ctx)
		if err != nil {
			logger.LogError("%s", err.Error())
			return &BuildResult{Success: false, Logs: logger.Logs}
		}
	}

	// Run the procfile provider to support apps that have a Procfile with a start command
	procfileProvider := &procfile.ProcfileProvider{}
	if _, err := procfileProvider.Plan(ctx); err != nil {
		logger.LogError("%s", err.Error())
		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	buildPlan, resolvedPackages, err := ctx.Generate()
	if err != nil {
		logger.LogError("%s", err.Error())
		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	// Error if there are no commands in the build plan
	var atLeastOneCommand = false
	for _, step := range buildPlan.Steps {
		if len(step.Commands) > 0 {
			atLeastOneCommand = true
		}
	}

	if !atLeastOneCommand {
		logger.LogError("%s", getNoProviderError(app))
		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	// Error if there is no start command and we are configured to error on it
	if options.ErrorMissingStartCommand && ctx.Deploy.StartCmd == "" {
		startCmdHelp := "No start command was found."
		if providerHelp := providerToUse.StartCommandHelp(); providerHelp != "" {
			startCmdHelp += "\n\n" + providerHelp
		}
		logger.LogError("%s", startCmdHelp)

		return &BuildResult{Success: false, Logs: logger.Logs}
	}

	buildResult := &BuildResult{
		RailpackVersion:   options.RailpackVersion,
		Plan:              buildPlan,
		ResolvedPackages:  resolvedPackages,
		Metadata:          ctx.Metadata.Properties,
		DetectedProviders: []string{detectedProviderName},
		Logs:              logger.Logs,
		Success:           true,
	}

	return buildResult
}

// GetConfig merges the options, environment, and file config into a single config
func GetConfig(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions, logger *logger.Logger) (*config.Config, error) {
	optionsConfig := GenerateConfigFromOptions(options)

	envConfig := GenerateConfigFromEnvironment(app, env)

	fileConfig, err := GenerateConfigFromFile(app, env, options, logger)
	if err != nil {
		return nil, err
	}

	mergedConfig := config.Merge(optionsConfig, envConfig, fileConfig)

	return mergedConfig, nil
}

// GenerateConfigFromFile generates a config from the config file
func GenerateConfigFromFile(app *app.App, env *app.Environment, options *GenerateBuildPlanOptions, logger *logger.Logger) (*config.Config, error) {
	configFileName := defaultConfigFileName
	if options.ConfigFilePath != "" {
		configFileName = options.ConfigFilePath
	}

	if envConfigFileName, _ := env.GetConfigVariable("CONFIG_FILE"); envConfigFileName != "" {
		configFileName = envConfigFileName
	}

	if !app.HasMatch(configFileName) {
		if configFileName != defaultConfigFileName {
			logger.LogWarn("Config file `%s` not found", configFileName)
		}

		return config.EmptyConfig(), nil
	}

	config := config.EmptyConfig()
	if err := app.ReadJSON(configFileName, config); err != nil {
		return nil, err
	}

	logger.LogInfo("Using config file `%s`", configFileName)

	fmt.Printf("CONFIG: %v\n", config)
	for _, step := range config.Steps {
		fmt.Printf("STEP: %v (nil? %v)\n", step.Secrets, step.Secrets == nil)
	}

	return config, nil
}

// GenerateConfigFromEnvironment generates a config from the environment
func GenerateConfigFromEnvironment(app *app.App, env *app.Environment) *config.Config {
	config := config.EmptyConfig()

	if env == nil {
		return config
	}

	if buildCmdVar, _ := env.GetConfigVariable("BUILD_CMD"); buildCmdVar != "" {
		buildStep := config.GetOrCreateStep("build")
		buildStep.Commands = []plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecShellCommand(buildCmdVar, plan.ExecOptions{CustomName: buildCmdVar}),
		}
	}

	if startCmdVar, _ := env.GetConfigVariable("START_CMD"); startCmdVar != "" {
		config.Deploy.StartCmd = startCmdVar
	}

	if envPackages, _ := env.GetConfigVariable("PACKAGES"); envPackages != "" {
		config.Packages = make(map[string]string)
		for _, pkg := range strings.Split(envPackages, " ") {
			// TODO: We should support specifying a version here (e.g. "node@18" or just "node")
			config.Packages[pkg] = "latest"
		}
	}

	if envAptPackages, _ := env.GetConfigVariable("BUILD_APT_PACKAGES"); envAptPackages != "" {
		config.BuildAptPackages = strings.Split(envAptPackages, " ")
	}

	if envAptPackages, _ := env.GetConfigVariable("DEPLOY_APT_PACKAGES"); envAptPackages != "" {
		config.Deploy.AptPackages = strings.Split(envAptPackages, " ")
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
		buildStep.Commands = []plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecShellCommand(options.BuildCommand, plan.ExecOptions{CustomName: options.BuildCommand}),
		}
	}

	if options.StartCommand != "" {
		config.Deploy.StartCmd = options.StartCommand
	}

	return config
}

func getProviders(ctx *generate.GenerateContext, config *config.Config) (providers.Provider, string) {
	allProviders := providers.GetLanguageProviders()

	var providerToUse providers.Provider
	var detectedProvider string

	// Even if there are providers manually specified, we want to detect to see what type of app this is
	for _, provider := range allProviders {
		matched, err := provider.Detect(ctx)
		if err != nil {
			log.Warnf("Failed to detect provider `%s`: %s", provider.Name(), err.Error())
			continue
		}

		if matched {
			detectedProvider = provider.Name()

			// If there are no providers manually specified in the config,
			if config.Provider == nil {
				if err := provider.Initialize(ctx); err != nil {
					ctx.Logger.LogWarn("Failed to initialize provider `%s`: %s", provider.Name(), err.Error())
					continue
				}

				ctx.Logger.LogInfo("Detected %s", utils.CapitalizeFirst(provider.Name()))

				providerToUse = provider
			}

			break
		}
	}

	if config.Provider != nil {
		provider := providers.GetProvider(*config.Provider)

		if provider == nil {
			ctx.Logger.LogWarn("Provider `%s` not found", *config.Provider)
			return providerToUse, detectedProvider
		}

		if err := provider.Initialize(ctx); err != nil {
			ctx.Logger.LogWarn("Failed to initialize provider `%s`: %s", *config.Provider, err.Error())
			return providerToUse, detectedProvider
		}

		ctx.Logger.LogInfo("Using provider %s from config", utils.CapitalizeFirst(*config.Provider))
		providerToUse = provider
	}

	return providerToUse, detectedProvider
}

func getNoProviderError(app *app.App) string {
	providerNames := []string{}
	for _, provider := range providers.GetLanguageProviders() {
		providerNames = append(providerNames, utils.CapitalizeFirst(provider.Name()))
	}

	files, _ := app.FindFiles("*")
	dirs, _ := app.FindDirectories("*")

	fileTree := "./\n"

	for i, dir := range dirs {
		prefix := "├── "
		if i == len(dirs)-1 && len(files) == 0 {
			prefix = "└── "
		}
		fileTree += fmt.Sprintf("%s%s/\n", prefix, dir)
	}

	for i, file := range files {
		prefix := "├── "
		if i == len(files)-1 {
			prefix = "└── "
		}
		fileTree += fmt.Sprintf("%s%s\n", prefix, file)
	}

	errorMsg := "Railpack could not determine how to build the app.\n\n"
	errorMsg += "The following languages are supported:\n"
	for _, provider := range providerNames {
		errorMsg += fmt.Sprintf("- %s\n", provider)
	}

	errorMsg += "\nThe app contents that Railpack analyzed contains:\n\n"
	errorMsg += fileTree
	errorMsg += "\n"
	errorMsg += "Check out the docs for more information: https://railpack.com"

	return errorMsg
}
