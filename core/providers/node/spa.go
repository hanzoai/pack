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
	DefaultDistDir       = "dist"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

func (p *NodeProvider) isSPA(ctx *generate.GenerateContext) bool {
	return p.isVite(ctx)
}

func (p *NodeProvider) DeploySPA(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	outputDir := p.getOutputDirectory(ctx)
	fmt.Println("outputDir", outputDir)

	data := map[string]interface{}{
		"DIST_DIR": path.Join("/app", outputDir),
	}

	caddyfileTemplate, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, data)
	if err != nil {
		return err
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
			Include: []string{"."},
			Exclude: []string{"node_modules"},
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
	}

	if outputDir == "" {
		outputDir = DefaultDistDir
	}

	return outputDir
}

func (p *NodeProvider) isReact() bool {
	return p.hasDependency("react")
}

func (p *NodeProvider) isVue() bool {
	return p.hasDependency("vue")
}

func (p *NodeProvider) isSvelte() bool {
	return p.hasDependency("svelte") && !p.hasDependency("@sveltejs/kit")
}

func (p *NodeProvider) isPreact() bool {
	return p.hasDependency("preact")
}

func (p *NodeProvider) isLit() bool {
	return p.hasDependency("lit")
}

func (p *NodeProvider) isSolidJs() bool {
	return p.hasDependency("solid-js")
}

func (p *NodeProvider) isQwik() bool {
	return p.hasDependency("@builder.io/qwik")
}
