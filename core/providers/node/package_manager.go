package node

import (
	"fmt"
	"strings"

	a "github.com/railwayapp/railpack-go/core/app"
	"github.com/railwayapp/railpack-go/core/generate"
	"github.com/railwayapp/railpack-go/core/plan"
)

const (
	PackageManagerNpm   PackageManager = "npm"
	PackageManagerPnpm  PackageManager = "pnpm"
	PackageManagerBun   PackageManager = "bun"
	PackageManagerYarn1 PackageManager = "yarn1"
	PackageManagerYarn2 PackageManager = "yarn2"
)

func (p PackageManager) Name() string {
	switch p {
	case PackageManagerNpm:
		return "npm"
	case PackageManagerPnpm:
		return "pnpm"
	case PackageManagerBun:
		return "bun"
	case PackageManagerYarn1, PackageManagerYarn2:
		return "yarn"
	default:
		return ""
	}
}

func (p PackageManager) RunCmd(cmd string) string {
	return fmt.Sprintf("%s run %s", p.Name(), cmd)
}

func (p PackageManager) RunScriptCommand(cmd string) string {
	if p == PackageManagerBun {
		return "bun " + cmd
	}
	return "node " + cmd
}

func (p PackageManager) installDependencies(ctx *generate.GenerateContext, packageJson *PackageJson, install *generate.CommandStepBuilder) {
	hasPostInstall := packageJson.Scripts != nil && packageJson.Scripts["postinstall"] != ""
	hasPrepare := packageJson.Scripts != nil && packageJson.Scripts["prepare"] != ""

	// If there is a postinstall script, we need the entire app to be copied
	// This is to handle things like patch-package
	if hasPostInstall || hasPrepare {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand(".", "."),
		})
	} else {
		for _, file := range p.SupportingInstallFiles(ctx.App) {
			install.AddCommands([]plan.Command{
				plan.NewCopyCommand(file, file),
			})
		}
	}

	p.InstallDeps(ctx, install)
}

func (p PackageManager) InstallDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	switch p {
	case PackageManagerNpm:
		hasLockfile := ctx.App.HasMatch("package-lock.json")
		npmCache := ctx.Caches.AddCache("npm-install", "/root/.npm")

		if hasLockfile {
			install.AddCommand(plan.NewExecCommand("npm ci", plan.ExecOptions{Caches: []string{npmCache}}))
		} else {
			install.AddCommand(plan.NewExecCommand("npm install",
				plan.ExecOptions{Caches: []string{npmCache}}))
		}
	case PackageManagerPnpm:
		install.AddCommand(plan.NewExecCommand("pnpm install --frozen-lockfile", plan.ExecOptions{
			Caches: []string{ctx.Caches.AddCache("pnpm-install", "/root/.local/share/pnpm/store/v3")},
		}))
	case PackageManagerBun:
		install.AddCommand(plan.NewExecCommand("bun i --no-save", plan.ExecOptions{
			Caches: []string{ctx.Caches.AddCache("bun-install", "/root/.bun/install/cache")},
		}))
	case PackageManagerYarn1:
		install.AddCommand(plan.NewExecCommand("yarn install --frozen-lockfile", plan.ExecOptions{
			Caches: []string{ctx.Caches.AddCache("yarn-install", "/usr/local/share/.cache/yarn")},
		}))
	case PackageManagerYarn2:
		install.AddCommand(plan.NewExecCommand("yarn install --check-cache", plan.ExecOptions{
			Caches: []string{ctx.Caches.AddCache("yarn-install", "/usr/local/share/.cache/yarn")},
		}))
	}
}

// SupportingInstallFiles returns a list of files that are needed to install dependencies
func (p PackageManager) SupportingInstallFiles(app *a.App) []string {
	patterns := []string{
		"**/package.json",
		"**/package-lock.json",
		"**/pnpm-workspace.yaml",
		"**/yarn.lock",
		"**/pnpm-lock.yaml",
		"**/bun.lockb",
		"**/bun.lock",
		"**/yarn.lock",
		"**/node_modules",
		"**/.pnp.*",        // Yarn Plug'n'Play files
		"**/.yarnrc.yml",   // Yarn 2+ config
		"**/.npmrc",        // NPM config
		"**/.node-version", // Node version file
		"**/.nvmrc",        // NVM config
	}

	var allFiles []string
	for _, pattern := range patterns {
		files, err := app.FindFiles(pattern)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles
}

// InstallPackages installs specific versions of package managers by analyzing the users code
func (p PackageManager) InstallPackages(ctx *generate.GenerateContext, packages *generate.MiseStepBuilder) {
	// NPM
	if p == PackageManagerNpm {
		npm := packages.Default("npm", "latest")

		lockfile, err := ctx.App.ReadFile("package-lock.json")
		if err != nil {
			lockfile = ""
		}

		if strings.Contains(lockfile, "\"lockfileVersion\": 1") {
			packages.Version(npm, "6", "package-lock.json")
		} else if strings.Contains(lockfile, "\"lockfileVersion\": 2") {
			packages.Version(npm, "8", "package-lock.json")
		} else if strings.Contains(lockfile, "\"lockfileVersion\": 3") {
			packages.Version(npm, "9", "package-lock.json")
		}
	}

	// Pnpm
	if p == PackageManagerPnpm {
		pnpm := packages.Default("pnpm", "latest")

		lockfile, err := ctx.App.ReadFile("pnpm-lock.yaml")
		if err == nil {
			if strings.HasPrefix(lockfile, "lockfileVersion: 5.3") {
				packages.Version(pnpm, "6", "pnpm-lock.yaml")
			} else if strings.HasPrefix(lockfile, "lockfileVersion: 5.4") {
				packages.Version(pnpm, "7", "pnpm-lock.yaml")
			} else if strings.HasPrefix(lockfile, "lockfileVersion: '6.0'") {
				packages.Version(pnpm, "8", "pnpm-lock.yaml")
			}
		}
	}

	// Yarn
	if p == PackageManagerYarn1 {
		packages.Default("yarn", "1")
		packages.AddAptPackage("tar")
		packages.AddAptPackage("gpg")
	} else if p == PackageManagerYarn2 {
		packages.Default("yarn", "2")
	}
}
