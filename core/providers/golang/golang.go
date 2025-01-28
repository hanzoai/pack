package golang

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/railwayapp/railpack-go/core/generate"
	"github.com/railwayapp/railpack-go/core/plan"
)

const (
	DEFAULT_GO_VERSION = "1.23"
	GO_BUILD_CACHE_KEY = "go-build"
	GO_BINARY_NAME     = "out"
	START_IMAGE        = "alpine:latest"
)

type GoProvider struct{}

func (p *GoProvider) Name() string {
	return "golang"
}

func (p *GoProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	isGo := p.isGoMod(ctx) || ctx.App.HasMatch("main.go")
	if !isGo {
		return false, nil
	}

	packages, err := p.Packages(ctx)
	if err != nil {
		return false, err
	}

	install, err := p.Install(ctx, packages)
	if err != nil {
		return false, err
	}

	_, err = p.Build(ctx, packages, install)
	if err != nil {
		return false, err
	}

	ctx.Start.Command = fmt.Sprintf("./%s", GO_BINARY_NAME)

	if !p.hasCGOEnabled(ctx) {
		ctx.Start.Paths = []string{GO_BINARY_NAME}

		ctx.Start.BaseImage = START_IMAGE
		if startImage, _ := ctx.Env.GetConfigVariable("START_IMAGE"); startImage != "" {
			ctx.Start.BaseImage = startImage
		}
	}

	if p.isGin(ctx) {
		ctx.Start.Env["GIN_MODE"] = "release"
	}

	p.addMetadata(ctx)

	return true, nil
}

func (p *GoProvider) Build(ctx *generate.GenerateContext, packages *generate.MiseStepBuilder, install *generate.CommandStepBuilder) (*generate.CommandStepBuilder, error) {
	var buildCmd string

	if binName, _ := ctx.Env.GetConfigVariable("GO_BIN"); binName != "" {
		// If there is a RAILPACK_GO_BIN variable, use that
		buildCmd = fmt.Sprintf("go build -o %s ./cmd/%s", GO_BINARY_NAME, binName)
	} else if p.isGoMod(ctx) && p.hasRootGoFiles(ctx) {
		buildCmd = fmt.Sprintf("go build -o %s", GO_BINARY_NAME)
	} else if dirs, err := ctx.App.FindDirectories("cmd/*"); err == nil && len(dirs) > 0 {
		// Try to find a command in the cmd directory
		cmdName := filepath.Base(dirs[0])
		buildCmd = fmt.Sprintf("go build -o %s ./cmd/%s", GO_BINARY_NAME, cmdName)
	} else if p.isGoMod(ctx) {
		buildCmd = fmt.Sprintf("go build -o %s", GO_BINARY_NAME)
	} else if ctx.App.HasMatch("main.go") {
		buildCmd = fmt.Sprintf("go build -o %s main.go", GO_BINARY_NAME)
	}

	if buildCmd == "" {
		return nil, nil
	}

	build := ctx.NewCommandStep("build")
	build.AddCommands([]plan.Command{
		plan.NewCopyCommand("."),
		plan.NewExecCommand(buildCmd, plan.ExecOptions{
			Caches: []string{p.goBuildCacheKey(ctx)},
		}),
	})

	if packages != nil {
		build.DependsOn = append(build.DependsOn, packages.DisplayName)
	}

	if install != nil {
		build.DependsOn = append(build.DependsOn, install.DisplayName)
	}

	return build, nil
}

func (p *GoProvider) Install(ctx *generate.GenerateContext, packages *generate.MiseStepBuilder) (*generate.CommandStepBuilder, error) {
	if !p.isGoMod(ctx) {
		return nil, nil
	}

	install := ctx.NewCommandStep("install")
	install.AddCommands([]plan.Command{
		plan.NewCopyCommand("go.mod"),
		plan.NewCopyCommand("go.sum"),
		plan.NewExecCommand("go mod download", plan.ExecOptions{
			Caches: []string{p.goBuildCacheKey(ctx)},
		}),
	})

	// If CGO is enabled, we need to install the gcc packages
	if p.hasCGOEnabled(ctx) {
		aptStep := ctx.NewAptStepBuilder("cgo")
		aptStep.Packages = []string{"gcc", "g++", "libc6-dev", "libgcc-9-dev", "libstdc++-9-dev"}
		install.DependsOn = append(install.DependsOn, aptStep.DisplayName)
	} else {
		install.AddCommand(plan.NewVariableCommand("CGO_ENABLED", "0"))
	}

	install.DependsOn = []string{packages.DisplayName}

	return install, nil
}

func (p *GoProvider) Packages(ctx *generate.GenerateContext) (*generate.MiseStepBuilder, error) {
	packages := ctx.GetMiseStepBuilder()

	goPkg := packages.Default("go", DEFAULT_GO_VERSION)

	if goModContents, err := ctx.App.ReadFile("go.mod"); err == nil {
		// Split content into lines and look for "go X.XX" line
		lines := strings.Split(string(goModContents), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "go ") {
				// Extract version number
				if goVersion := strings.TrimSpace(strings.TrimPrefix(line, "go")); goVersion != "" {
					packages.Version(goPkg, goVersion, "go.mod")
					break
				}
			}
		}
	}

	if envVersion, varName := ctx.Env.GetConfigVariable("GO_VERSION"); envVersion != "" {
		packages.Version(goPkg, envVersion, varName)
	}

	return packages, nil
}

func (p *GoProvider) addMetadata(ctx *generate.GenerateContext) {
	ctx.Metadata.Set("hasGoMod", strconv.FormatBool(p.isGoMod(ctx)))
	ctx.Metadata.Set("hasRootGoFiles", strconv.FormatBool(p.hasRootGoFiles(ctx)))
	ctx.Metadata.Set("hasGin", strconv.FormatBool(p.isGin(ctx)))
	ctx.Metadata.Set("hasCGOEnabled", strconv.FormatBool(p.hasCGOEnabled(ctx)))
}

func (p *GoProvider) goBuildCacheKey(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(GO_BUILD_CACHE_KEY, "/root/.cache/go-build")
}

func (p *GoProvider) hasRootGoFiles(ctx *generate.GenerateContext) bool {
	if files, err := ctx.App.FindFiles("*.go"); err == nil {
		for _, file := range files {
			if filepath.Dir(file) == "." {
				return true
			}
		}
	}
	return false
}

func (p *GoProvider) isGin(ctx *generate.GenerateContext) bool {
	if goModContents, err := ctx.App.ReadFile("go.mod"); err == nil {
		return strings.Contains(string(goModContents), "github.com/gin-gonic/gin")
	}

	return false
}

func (p *GoProvider) hasCGOEnabled(ctx *generate.GenerateContext) bool {
	return ctx.Env.GetVariable("CGO_ENABLED") == "1"
}

func (p *GoProvider) isGoMod(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("go.mod")
}
