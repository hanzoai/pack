package node

import (
	"fmt"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type PackageJson struct {
	Scripts        map[string]string `json:"scripts"`
	PackageManager *string           `json:"packageManager"`
	Engines        map[string]string `json:"engines"`
	Main           *string           `json:"main"`
}

type PackageManager string

const (
	DEFAULT_NODE_VERSION = "22"
	DEFAULT_BUN_VERSION  = "latest"
)

type NodeProvider struct{}

func (p *NodeProvider) Name() string {
	return "node"
}

func (p *NodeProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	packageJson, err := p.GetPackageJson(ctx.App)
	if err != nil {
		return false, err
	}

	if packageJson == nil {
		return false, nil
	}

	packages, err := p.Packages(ctx, packageJson)
	if err != nil {
		return false, err
	}

	install, err := p.Install(ctx, packages, packageJson)
	if err != nil {
		return false, err
	}

	if _, err := p.Build(ctx, install, packageJson); err != nil {
		return false, err
	}

	if err := p.start(ctx, packageJson); err != nil {
		return false, err
	}

	return true, nil
}

func (p *NodeProvider) start(ctx *generate.GenerateContext, packageJson *PackageJson) error {
	packageManager := p.getPackageManager(ctx.App)

	if start := p.getScripts(packageJson, "start"); start != "" {
		ctx.Start.Command = packageManager.RunCmd("start")
	} else if main := packageJson.Main; main != nil {
		ctx.Start.Command = packageManager.RunScriptCommand(*main)
	} else if files, err := ctx.App.FindFiles("{index.js,index.ts}"); err == nil && len(files) > 0 {
		ctx.Start.Command = packageManager.RunScriptCommand(files[0])
	}

	ctx.Start.Paths = append(ctx.Start.Paths, ".")

	return nil
}

func (p *NodeProvider) Build(ctx *generate.GenerateContext, install *generate.CommandStepBuilder, packageJson *PackageJson) (*generate.CommandStepBuilder, error) {
	packageManager := p.getPackageManager(ctx.App)
	_, ok := packageJson.Scripts["build"]
	if ok {
		build := ctx.NewCommandStep("build")

		build.AddCommands([]plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecCommand(packageManager.RunCmd("build")),
		})

		build.DependsOn = []string{install.DisplayName}

		return build, nil
	}

	return nil, nil
}

func (p *NodeProvider) Install(ctx *generate.GenerateContext, packages *generate.MiseStepBuilder, packageJson *PackageJson) (*generate.CommandStepBuilder, error) {
	var corepackStepName string
	if p.usesCorepack(packageJson) {
		corepackStep := ctx.NewCommandStep("corepack")
		corepackStep.AddCommands([]plan.Command{
			plan.NewCopyCommand("package.json"),
			plan.NewExecCommand("npm install -g corepack"),
			plan.NewExecCommand("corepack enable"),
			plan.NewExecCommand("corepack prepare --activate"),
		})
		corepackStepName = corepackStep.DisplayName
	}

	pkgManager := p.getPackageManager(ctx.App)

	install := ctx.NewCommandStep("install")
	install.DependsOn = []string{packages.DisplayName}
	pkgManager.installDependencies(ctx, packageJson, install)

	if corepackStepName != "" {
		install.DependsOn = []string{corepackStepName}
	}

	return install, nil
}

func (p *NodeProvider) Packages(ctx *generate.GenerateContext, packageJson *PackageJson) (*generate.MiseStepBuilder, error) {
	packageManager := p.getPackageManager(ctx.App)
	ctx.Metadata.Set("packageManager", string(packageManager))

	packages := ctx.GetMiseStepBuilder()

	// Node
	node := packages.Default("node", DEFAULT_NODE_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("NODE_VERSION"); envVersion != "" {
		packages.Version(node, envVersion, varName)
	}

	if packageJson.Engines != nil && packageJson.Engines["node"] != "" {
		packages.Version(node, packageJson.Engines["node"], "package.json > engines > node")
	}

	if packageManager == PackageManagerBun {
		bun := packages.Default("bun", DEFAULT_BUN_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("BUN_VERSION"); envVersion != "" {
			packages.Version(bun, envVersion, varName)
		}
	}

	packageManager.InstallPackages(ctx, packages)

	return packages, nil
}

func (p *NodeProvider) usesCorepack(packageJson *PackageJson) bool {
	return packageJson.PackageManager != nil
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
	if !app.HasMatch("package.json") {
		return nil, nil
	}

	var packageJson PackageJson
	err := app.ReadJSON("package.json", &packageJson)
	if err != nil {
		return nil, fmt.Errorf("error reading package.json: %w", err)
	}

	return &packageJson, nil
}

func (p *NodeProvider) getScripts(packageJson *PackageJson, name string) string {
	if scripts := packageJson.Scripts; scripts != nil {
		if script, ok := scripts[name]; ok {
			return script
		}
	}

	return ""
}
