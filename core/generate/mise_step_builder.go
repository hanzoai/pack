package generate

import (
	"fmt"
	"strings"

	a "github.com/railwayapp/railpack-go/core/app"
	"github.com/railwayapp/railpack-go/core/mise"
	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/railwayapp/railpack-go/core/resolver"
)

const (
	MisePackageStepName = "packages:mise"
)

type MiseStepBuilder struct {
	DisplayName           string
	Resolver              *resolver.Resolver
	SupportingAptPackages []string
	MisePackages          []*resolver.PackageRef
	SupportingMiseFiles   []string
	Assets                map[string]string
	DependsOn             []string
	Outputs               []string

	app *a.App
	env *a.Environment
}

func (c *GenerateContext) newMiseStepBuilder() *MiseStepBuilder {
	step := &MiseStepBuilder{
		DisplayName:           c.GetStepName(MisePackageStepName),
		Resolver:              c.resolver,
		MisePackages:          []*resolver.PackageRef{},
		SupportingAptPackages: []string{},
		Assets:                map[string]string{},
		DependsOn:             []string{},
		Outputs:               []string{"/mise/shims", "/mise/installs", "/usr/local/bin/mise", "/etc/mise/config.toml", "/root/.local/state/mise"},
		app:                   c.App,
		env:                   c.Env,
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *MiseStepBuilder) AddAptPackage(name string) {
	b.SupportingAptPackages = append(b.SupportingAptPackages, name)
}

func (b *MiseStepBuilder) Default(name string, defaultVersion string) resolver.PackageRef {
	for _, pkg := range b.MisePackages {
		if pkg.Name == name {
			return *pkg
		}
	}

	pkg := b.Resolver.Default(name, defaultVersion)
	b.MisePackages = append(b.MisePackages, &pkg)
	return pkg
}

func (b *MiseStepBuilder) Version(name resolver.PackageRef, version string, source string) {
	b.Resolver.Version(name, version, source)
}

func (b *MiseStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)

	step.DependsOn = b.DependsOn

	if _, ok := options.Caches["apt"]; !ok {
		aptCache := plan.NewCache("/var/cache/apt")
		aptCache.Type = plan.CacheTypeLocked
		options.Caches["apt"] = aptCache
	}

	if _, ok := options.Caches["mise"]; !ok {
		options.Caches["mise"] = plan.NewCache("/mise/cache")
	}

	// Install mise
	if len(b.MisePackages) > 0 {
		step.AddCommands([]plan.Command{
			plan.NewVariableCommand("MISE_DATA_DIR", "/mise"),
			plan.NewVariableCommand("MISE_CONFIG_DIR", "/mise"),
			plan.NewVariableCommand("MISE_INSTALL_PATH", "/usr/local/bin/mise"),
			plan.NewVariableCommand("MISE_CACHE_DIR", "/mise/cache"),
			plan.NewPathCommand("/mise/shims"),
			plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*'", plan.ExecOptions{
				CustomName: "install curl",
				CacheKey:   "apt",
			}),
			plan.NewExecCommand("sh -c 'curl -fsSL https://mise.run | sh'",
				plan.ExecOptions{
					CustomName: "install mise",
					CacheKey:   "mise",
				}),
		})

		// Add user mise config files if they exist
		supportingMiseConfigFiles := b.GetSupportingMiseConfigFiles(b.app.Source)
		for _, file := range supportingMiseConfigFiles {
			step.AddCommands([]plan.Command{
				plan.NewCopyCommand(file, "/app/"+file),
			})
		}
	}

	// Setup apt commands
	if len(b.SupportingAptPackages) > 0 {
		pkgString := strings.Join(b.SupportingAptPackages, " ")
		step.AddCommands([]plan.Command{
			plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+pkgString+" && rm -rf /var/lib/apt/lists/*'", plan.ExecOptions{
				CustomName: "install apt packages: " + pkgString,
				CacheKey:   "apt",
			}),
		})
	}

	// Setup mise commands
	if len(b.MisePackages) > 0 {
		packagesToInstall := make(map[string]string)
		for _, pkg := range b.MisePackages {
			resolved, ok := options.ResolvedPackages[pkg.Name]
			if ok && resolved.ResolvedVersion != nil {
				packagesToInstall[pkg.Name] = *resolved.ResolvedVersion
			}
		}

		miseToml, err := mise.GenerateMiseToml(packagesToInstall)
		if err != nil {
			return nil, fmt.Errorf("failed to generate mise.toml: %w", err)
		}

		b.Assets["mise.toml"] = miseToml

		pkgNames := make([]string, 0, len(packagesToInstall))
		for k := range packagesToInstall {
			pkgNames = append(pkgNames, k)
		}

		step.AddCommands([]plan.Command{
			plan.NewFileCommand("/etc/mise/config.toml", "mise.toml", plan.FileOptions{
				CustomName: "create mise config",
			}),
			plan.NewExecCommand("sh -c 'mise trust -a && mise install'", plan.ExecOptions{
				CustomName: "install mise packages: " + strings.Join(pkgNames, ", "),
				CacheKey:   "mise",
			}),
		})
	}

	step.Assets = b.Assets
	step.Outputs = b.Outputs

	return step, nil
}

var miseConfigFiles = []string{
	"mise.toml",
	".python-version",
	".nvmrc",
}

func (b *MiseStepBuilder) GetSupportingMiseConfigFiles(path string) []string {
	files := []string{}

	for _, file := range miseConfigFiles {
		if b.app.HasMatch(file) {
			files = append(files, file)
		}
	}

	return files
}
