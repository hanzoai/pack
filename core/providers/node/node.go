package node

import (
	"fmt"
	"maps"
	"path"
	"strings"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type PackageManager string

const (
	DEFAULT_NODE_VERSION = "22"
	DEFAULT_BUN_VERSION  = "latest"
)

type NodeProvider struct {
	packageJson    *PackageJson
	packageManager PackageManager
	workspace      *Workspace
}

func (p *NodeProvider) Name() string {
	return "node"
}

func (p *NodeProvider) Initialize(ctx *generate.GenerateContext) error {
	packageJson, err := p.GetPackageJson(ctx.App)
	if err != nil {
		return err
	}
	p.packageJson = packageJson

	p.packageManager = p.getPackageManager(ctx.App)

	workspace, err := NewWorkspace(ctx.App)
	if err != nil {
		return err
	}
	p.workspace = workspace

	return nil
}

func (p *NodeProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("package.json"), nil
}

func (p *NodeProvider) Plan(ctx *generate.GenerateContext) error {
	if p.packageJson == nil {
		return fmt.Errorf("package.json not loaded, did you call Initialize?")
	}

	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	// Install
	install := ctx.NewCommandStep("install")

	install.Inputs = []plan.StepInput{
		plan.NewStepInput(miseStep.Name()),
	}

	maps.Copy(install.Variables, p.GetNodeEnvVars(ctx))
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"NODE", "NPM", "BUN", "PNPM", "YARN", "CI"})
	install.AddPaths([]string{"/app/node_modules/.bin"})

	if p.usesCorepack() {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("package.json"),
			plan.NewExecCommand("corepack enable"),
			plan.NewExecCommand("corepack prepare --activate"),
		})
	}

	p.packageManager.installDependencies(ctx, p.packageJson, install)

	// Prune
	pruneStep := ctx.NewCommandStep("prune")
	pruneStep.AddCommands([]plan.Command{
		p.packageManager.PruneCommand(),
	})
	pruneStep.Variables["NPM_CONFIG_PRODUCTION"] = "true"
	pruneStep.Inputs = []plan.StepInput{
		plan.NewStepInput(install.Name()),
	}
	pruneStep.Secrets = []string{}

	// Build
	build := ctx.NewCommandStep("build")
	build.Inputs = []plan.StepInput{
		plan.NewStepInput(install.Name()),
	}
	_, ok := p.packageJson.Scripts["build"]
	if ok {
		build.AddCommands([]plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecCommand(p.packageManager.RunCmd("build")),
		})
	}

	ctx.Deploy.Inputs = append(ctx.Deploy.Inputs, []plan.StepInput{
		plan.NewImageInput(plan.RAILPACK_RUNTIME_IMAGE),
		plan.NewLocalInput("."),
		plan.NewStepInput(miseStep.Name(), plan.InputOptions{
			Include: miseStep.GetOutputPaths(),
		}),
		plan.NewStepInput(build.Name(), plan.InputOptions{
			Include: []string{"/app"},
			Exclude: []string{"node_modules"},
		}),
		plan.NewStepInput(pruneStep.Name(), plan.InputOptions{
			Include: []string{"/app/node_modules"}, // we only wanted the pruned node_modules
		}),
	}...)

	ctx.Deploy.StartCmd = "node --version"
	maps.Copy(ctx.Deploy.Variables, p.GetNodeEnvVars(ctx))

	// packages, err := p.Packages(ctx, p.packageJson)
	// if err != nil {
	// 	return err
	// }

	// install, err := p.Install(ctx, packages, p.packageJson)
	// if err != nil {
	// 	return err
	// }

	// if _, err := p.Build(ctx, install, p.packageJson); err != nil {
	// 	return err
	// }

	// if err := p.start(ctx, p.packageJson); err != nil {
	// 	return err
	// }

	return nil
}

func (p *NodeProvider) start(ctx *generate.GenerateContext, packageJson *PackageJson) error {
	packageManager := p.getPackageManager(ctx.App)

	if start := p.getScripts(packageJson, "start"); start != "" {
		ctx.Start.Command = packageManager.RunCmd("start")
	} else if main := packageJson.Main; main != "" {
		ctx.Start.Command = packageManager.RunScriptCommand(main)
	} else if files, err := ctx.App.FindFiles("{index.js,index.ts}"); err == nil && len(files) > 0 {
		ctx.Start.Command = packageManager.RunScriptCommand(files[0])
	}

	ctx.Start.AddOutputs([]string{"."})
	ctx.Start.AddEnvVars(p.GetNodeEnvVars(ctx))

	return nil
}

func (p *NodeProvider) Build(ctx *generate.GenerateContext, install *generate.CommandStepBuilder, packageJson *PackageJson) (*generate.CommandStepBuilder, error) {
	packageManager := p.getPackageManager(ctx.App)
	build := ctx.NewCommandStep("build")
	// build.DependsOn = []string{install.DisplayName}

	_, ok := packageJson.Scripts["build"]
	if ok {
		build.AddCommands([]plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecCommand(packageManager.RunCmd("build")),
		})

	}

	// Generic node_modules cache
	build.AddCache(ctx.Caches.AddCache("node-modules", "/app/node_modules/.cache"))

	// Add caches for Next.JS apps
	if nextApps, err := p.getNextApps(ctx); err == nil {
		ctx.Metadata.SetBool("nextjs", len(nextApps) > 0)

		for _, nextApp := range nextApps {
			nextCacheDir := path.Join("/app", nextApp, ".next/cache")
			build.AddCache(ctx.Caches.AddCache(fmt.Sprintf("next-%s", nextApp), nextCacheDir))
		}
	}

	return build, nil
}

func (p *NodeProvider) InstallNodeDeps(ctx *generate.GenerateContext) (*generate.CommandStepBuilder, error) {
	var corepackStepName string
	if p.usesCorepack() {
		corepackStep := ctx.NewCommandStep("corepack")
		corepackStep.AddCommands([]plan.Command{
			plan.NewCopyCommand("package.json"),
			plan.NewExecCommand("corepack enable"),
			plan.NewExecCommand("corepack prepare --activate"),
		})
		corepackStepName = corepackStep.DisplayName
		corepackStep.Secrets = []string{} // Don't include any secrets in this step

		// corepackStep.DependsOn = append(corepackStep.DependsOn, setup.DisplayName)
	}

	pkgManager := p.getPackageManager(ctx.App)

	install := ctx.NewCommandStep("install")
	maps.Copy(install.Variables, p.GetNodeEnvVars(ctx))

	// install.DependsOn = append(install.DependsOn, []string{packages.DisplayName, setup.DisplayName}...)

	// We only want to invalidate the install step when these secrets change, not all of them
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"NODE", "NPM", "BUN", "PNPM", "YARN", "CI"})
	install.AddPaths([]string{"/app/node_modules/.bin"})

	pkgManager.installDependencies(ctx, p.packageJson, install)

	if corepackStepName != "" {
		// install.DependsOn = append(install.DependsOn, corepackStepName)
	}

	return nil, nil
}

func (p *NodeProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	// Node
	if p.packageManager.requiresNode(p.packageJson) {
		node := miseStep.Default("node", DEFAULT_NODE_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("NODE_VERSION"); envVersion != "" {
			miseStep.Version(node, envVersion, varName)
		}

		if p.packageJson.Engines != nil && p.packageJson.Engines["node"] != "" {
			miseStep.Version(node, p.packageJson.Engines["node"], "package.json > engines > node")
		}
	}

	// Bun
	if p.packageManager == PackageManagerBun {
		bun := miseStep.Default("bun", DEFAULT_BUN_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("BUN_VERSION"); envVersion != "" {
			miseStep.Version(bun, envVersion, varName)
		}
	}

	p.packageManager.GetPackageManagerPackages(ctx, miseStep)

	if p.usesCorepack() {
		miseStep.Variables["MISE_NODE_COREPACK"] = "true"
	}
}

func (p *NodeProvider) GetNodeEnvVars(ctx *generate.GenerateContext) map[string]string {
	envVars := map[string]string{
		"NODE_ENV":                   "production",
		"NPM_CONFIG_PRODUCTION":      "false",
		"NPM_CONFIG_UPDATE_NOTIFIER": "false",
		"NPM_CONFIG_FUND":            "false",
		"YARN_PRODUCTION":            "false",
		"CI":                         "true",
	}

	return envVars
}

func (p *NodeProvider) usesCorepack() bool {
	return p.packageJson.PackageManager != nil
}

func (p *NodeProvider) getPackageManager(app *app.App) PackageManager {
	packageManager := PackageManagerNpm

	if app.HasMatch("pnpm-lock.yaml") {
		packageManager = PackageManagerPnpm
	} else if app.HasMatch("bun.lockb") || app.HasMatch("bun.lock") {
		packageManager = PackageManagerBun
	} else if app.HasMatch(".yarnrc.yml") || app.HasMatch(".yarnrc.yaml") {
		packageManager = PackageManagerYarn2
	} else if app.HasMatch("yarn.lock") {
		packageManager = PackageManagerYarn1
	}

	return packageManager
}

func (p *NodeProvider) GetPackageJson(app *app.App) (*PackageJson, error) {
	packageJson := NewPackageJson()
	if !app.HasMatch("package.json") {
		return packageJson, nil
	}

	err := app.ReadJSON("package.json", packageJson)
	if err != nil {
		return nil, fmt.Errorf("error reading package.json: %w", err)
	}

	return packageJson, nil
}

func (p *NodeProvider) getScripts(packageJson *PackageJson, name string) string {
	if scripts := packageJson.Scripts; scripts != nil {
		if script, ok := scripts[name]; ok {
			return script
		}
	}

	return ""
}

func (p *NodeProvider) getNextApps(ctx *generate.GenerateContext) ([]string, error) {
	nextPaths, err := p.filterPackageJson(ctx, func(packageJson *PackageJson) bool {
		if packageJson.HasScript("build") {
			return strings.Contains(packageJson.Scripts["build"], "next build")
		}

		return false
	})
	if err != nil {
		return nil, err
	}

	return nextPaths, nil
}

func (p *NodeProvider) filterPackageJson(ctx *generate.GenerateContext, filterFunc func(packageJson *PackageJson) bool) ([]string, error) {
	filteredPaths := []string{}

	files, err := ctx.App.FindFiles("**/package.json")
	if err != nil {
		return filteredPaths, err
	}

	for _, file := range files {
		var packageJson PackageJson
		err := ctx.App.ReadJSON(file, &packageJson)
		if err != nil {
			return filteredPaths, err
		}

		if filterFunc(&packageJson) {
			dirPath := strings.TrimSuffix(file, "package.json")
			filteredPaths = append(filteredPaths, dirPath)
		}
	}

	return filteredPaths, nil
}
