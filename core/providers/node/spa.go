package node

import (
	_ "embed"
	"fmt"
	"path"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DefaultCaddyfilePath = "/Caddyfile"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

func (p *NodeProvider) isSPA(ctx *generate.GenerateContext) bool {
	if ctx.Env.IsConfigVariableTruthy("NO_SPA") {
		ctx.Logger.LogInfo("Skipping SPA deployment because NO_SPA is set")
		return false
	}

	isVite := p.isVite(ctx)
	isAstro := p.isAstroSPA(ctx)
	isCRA := p.isCRA(ctx)

	return (isVite || isAstro || isCRA) && p.getOutputDirectory(ctx) != ""
}

func (p *NodeProvider) getSPAFramework(ctx *generate.GenerateContext) string {
	if p.isVite(ctx) {
		return "vite"
	} else if p.isAstro(ctx) {
		return "astro"
	} else if p.isCRA(ctx) {
		return "CRA"
	}

	return ""
}

func (p *NodeProvider) DeploySPA(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	outputDir := p.getOutputDirectory(ctx)
	spaFramework := p.getSPAFramework(ctx)

	ctx.Logger.LogInfo("Deploying as %s static site", spaFramework)
	ctx.Logger.LogInfo("Output directory: %s", outputDir)

	data := map[string]interface{}{
		"DIST_DIR": path.Join("/app", outputDir),
	}

	caddyfileTemplate, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, data)
	if err != nil {
		return err
	}

	if caddyfileTemplate.Filename != "" {
		ctx.Logger.LogInfo("Using custom Caddyfile: %s", caddyfileTemplate.Filename)
	}

	installCaddyStep := ctx.NewInstallBinStepBuilder("packages:caddy")
	installCaddyStep.Default("caddy", "latest")

	caddy := ctx.NewCommandStep("caddy")
	caddy.AddInput(plan.NewStepInput(installCaddyStep.Name()))
	caddy.AddCommands([]plan.Command{
		plan.NewFileCommand(DefaultCaddyfilePath, "Caddyfile"),
		plan.NewExecCommand(fmt.Sprintf("caddy fmt --overwrite %s", DefaultCaddyfilePath)),
	})
	caddy.Assets = map[string]string{
		"Caddyfile": caddyfileTemplate.Contents,
	}

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", DefaultCaddyfilePath)

	ctx.Deploy.Inputs = []plan.Input{
		ctx.DefaultRuntimeInput(),
		plan.NewStepInput(installCaddyStep.Name(), plan.InputOptions{
			Include: installCaddyStep.GetOutputPaths(),
		}),
		plan.NewStepInput(build.Name(), plan.InputOptions{
			Include: []string{outputDir},
		}),
		plan.NewStepInput(caddy.Name(), plan.InputOptions{
			Include: []string{DefaultCaddyfilePath},
		}),
		plan.NewLocalInput("."),
	}

	return nil
}

func (p *NodeProvider) getOutputDirectory(ctx *generate.GenerateContext) string {
	outputDir := ""

	if dir, _ := ctx.Env.GetConfigVariable("SPA_OUTPUT_DIR"); dir != "" {
		outputDir = dir
	} else if p.isVite(ctx) {
		outputDir = p.getViteOutputDirectory(ctx)
	} else if p.isAstroSPA(ctx) {
		outputDir = p.getAstroOutputDirectory(ctx)
	} else if p.isCRA(ctx) {
		outputDir = p.getCRAOutputDirectory(ctx)
	}

	return outputDir
}

// func (p *NodeProvider) isReact() bool {
// 	return p.hasDependency("react")
// }

// func (p *NodeProvider) isVue() bool {
// 	return p.hasDependency("vue")
// }

// func (p *NodeProvider) isSvelte() bool {
// 	return p.hasDependency("svelte") && !p.hasDependency("@sveltejs/kit")
// }

// func (p *NodeProvider) isPreact() bool {
// 	return p.hasDependency("preact")
// }

// func (p *NodeProvider) isLit() bool {
// 	return p.hasDependency("lit")
// }

// func (p *NodeProvider) isSolidJs() bool {
// 	return p.hasDependency("solid-js")
// }

// func (p *NodeProvider) isQwik() bool {
// 	return p.hasDependency("@builder.io/qwik")
// }
