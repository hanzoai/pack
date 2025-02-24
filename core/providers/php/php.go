package php

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/stretchr/objx"
)

const (
	DEFAULT_PHP_VERSION = "8.4.3"
)

type PhpProvider struct{}

func (p *PhpProvider) Name() string {
	return "php"
}

func (p *PhpProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("index.php") ||
		ctx.App.HasMatch("composer.json"), nil
}

func (p *PhpProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *PhpProvider) Plan(ctx *generate.GenerateContext) error {
	phpImageStep, err := p.phpImagePackage(ctx)
	if err != nil {
		return err
	}

	// Nginx
	nginx := ctx.NewCommandStep("nginx")
	nginx.AddInput(plan.NewStepInput(phpImageStep.Name()))
	nginx.AddCommands([]plan.Command{
		plan.NewFileCommand("/etc/nginx/railpack.conf", "nginx.conf", plan.FileOptions{CustomName: "create nginx config"}),
		plan.NewExecCommand("nginx -t -c /etc/nginx/railpack.conf"),
		plan.NewFileCommand("/etc/php-fpm.conf", "php-fpm.conf", plan.FileOptions{CustomName: "create php-fpm config"}),
		plan.NewFileCommand("/start-nginx.sh", "start-nginx.sh", plan.FileOptions{
			CustomName: "create start nginx script",
			Mode:       0755,
		}),
	})

	nginx.Assets["start-nginx.sh"] = startNginxScriptAsset
	configFiles, err := p.getConfigFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config files: %w", err)
	}

	nginx.Assets["nginx.conf"] = configFiles.NginxConf
	nginx.Assets["nginx.conf"] = configFiles.NginxConf
	nginx.Assets["php-fpm.conf"] = configFiles.PhpFpmConf

	// Composer
	composer := ctx.NewCommandStep("install:composer")
	composer.AddInput(plan.NewStepInput(nginx.Name()))
	if _, err := p.readComposerJson(ctx); err == nil {
		composer.AddCache(ctx.Caches.AddCache("composer", "/opt/cache/composer"))
		composer.AddEnvVars(map[string]string{"COMPOSER_CACHE_DIR": "/opt/cache/composer"})

		composer.AddCommands([]plan.Command{
			// Copy composer from the composer image
			plan.CopyCommand{Image: "composer:latest", Src: "/usr/bin/composer", Dest: "/usr/bin/composer"},
			plan.NewCopyCommand("."),
			plan.NewExecCommand("composer install --ignore-platform-reqs"),
		})
	}

	// Node (if necessary)
	nodeProvider := node.NodeProvider{}
	isNode, err := nodeProvider.Detect(ctx)
	if err != nil {
		return err
	}

	if isNode {
		nodeProvider.Initialize(ctx)

		miseStep := ctx.GetMiseStepBuilder()
		nodeProvider.InstallMisePackages(ctx, miseStep)

		install := ctx.NewCommandStep("install:node")
		install.AddInput(plan.NewStepInput(miseStep.Name()))
		nodeProvider.InstallNodeDeps(ctx, install)

		prune := ctx.NewCommandStep("prune:node")
		prune.AddInput(plan.NewStepInput(install.Name()))
		nodeProvider.PruneNodeDeps(ctx, prune)

		build := ctx.NewCommandStep("build")
		build.Inputs = []plan.Input{
			plan.NewStepInput(composer.Name()),

			// Include the app and mise packages (node, pnpm, etc.)
			plan.NewStepInput(install.Name(), plan.InputOptions{
				Include: append([]string{"."}, miseStep.GetOutputPaths()...),
			}),
		}
		nodeProvider.Build(ctx, build)

		ctx.Deploy.Inputs = []plan.Input{
			plan.NewStepInput(composer.Name()),
			plan.NewStepInput(build.Name(), plan.InputOptions{
				Include: []string{"."},
				Exclude: []string{"node_modules"},
			}),
			plan.NewStepInput(prune.Name(), plan.InputOptions{
				Include: []string{"/app/node_modules"},
			}),
			plan.NewLocalInput("."),
		}
	} else {
		// A manual build command will go here
		build := ctx.NewCommandStep("build")
		build.AddInput(plan.NewStepInput(composer.Name()))

		ctx.Deploy.Inputs = []plan.Input{
			plan.NewStepInput(build.Name()),
			plan.NewLocalInput("."),
		}
	}

	ctx.Deploy.StartCmd = "bash /start-nginx.sh"

	if p.usesLaravel(ctx) {
		ctx.Deploy.Variables["IS_LARAVEL"] = "true"
	}

	return nil
}

func (p *PhpProvider) usesLaravel(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("artisan")
}

type ConfigFiles struct {
	NginxConf  string
	PhpFpmConf string
}

func (p *PhpProvider) getConfigFiles(ctx *generate.GenerateContext) (*ConfigFiles, error) {
	phpRootDir := "/app"
	if variable := ctx.Env.GetVariable("RAILPACK_PHP_ROOT_DIR"); variable != "" {
		phpRootDir = variable
	} else if p.usesLaravel(ctx) {
		phpRootDir = "/app/public"
	}

	data := map[string]interface{}{
		"RAILPACK_PHP_ROOT_DIR": phpRootDir,
		"IS_LARAVEL":            p.usesLaravel(ctx),
	}

	nginxConf, err := readFileOrTemplateWithDefault(ctx, "nginx.conf", "nginx.template.conf", nginxConfTemplateAsset, data)
	if err != nil {
		return nil, err
	}

	phpFpmConf, err := readFileOrTemplateWithDefault(ctx, "php-fpm.conf", "php-fpm.template.conf", phpFpmConfTemplateAsset, data)
	if err != nil {
		return nil, err
	}

	return &ConfigFiles{
		NginxConf:  nginxConf,
		PhpFpmConf: phpFpmConf,
	}, nil
}

func (p *PhpProvider) phpImagePackage(ctx *generate.GenerateContext) (*generate.ImageStepBuilder, error) {
	imageStep := ctx.NewImageStep("packages:image", func(options *generate.BuildStepOptions) string {
		if phpVersion, ok := options.ResolvedPackages["php"]; ok && phpVersion.ResolvedVersion != nil {
			return getPhpImage(*phpVersion.ResolvedVersion)
		}

		// Return the default if we were not able to resolve the version
		return getPhpImage(DEFAULT_PHP_VERSION)
	})

	imageStep.AptPackages = append(imageStep.AptPackages, "nginx", "git", "zip", "unzip")

	// Include both build and runtime apt packages since we don't have a separate runtime image
	imageStep.AptPackages = append(imageStep.AptPackages, ctx.Config.BuildAptPackages...)
	imageStep.AptPackages = append(imageStep.AptPackages, ctx.Config.Deploy.AptPackages...)

	php := imageStep.Default("php", DEFAULT_PHP_VERSION)

	// Read composer.json to get the PHP version
	if composerJson, err := p.readComposerJson(ctx); err == nil {
		phpVersion := objx.New(composerJson).Get("require.php")
		if phpVersion.IsStr() {
			if strings.HasPrefix(phpVersion.Str(), "^") {
				imageStep.Version(php, strings.TrimPrefix(phpVersion.Str(), "^"), "composer.json > require > php")
			} else {
				imageStep.Version(php, phpVersion.Str(), "composer.json > require > php")
			}
		}
	}

	return imageStep, nil
}

// readFileOrTemplateWithDefault reads a file or template from the app, or fallsback to a static default
// the template is then rendered with the data and returned
func readFileOrTemplateWithDefault(ctx *generate.GenerateContext, filename string, templateFilename string, defaultContents string, data map[string]interface{}) (string, error) {
	var conf string
	var confTemplate string

	// The user has a custom nginx.conf file that we should use
	if userConf, err := ctx.App.ReadFile(filename); err == nil {
		conf = userConf
	}

	// The user has a custom nginx.template.conf file that we should use
	if userTemplateConf, err := ctx.App.ReadFile(templateFilename); err == nil {
		confTemplate = userTemplateConf
	} else {
		// Otherwise, use the default nginx.template.conf file
		confTemplate = defaultContents
	}

	// We need to render the nginx.conf template if the user has a custom one
	if conf == "" && confTemplate != "" {
		tmpl, err := template.New(filename).Parse(confTemplate)
		if err != nil {
			return "", fmt.Errorf("failed to parse %s template: %w", filename, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute nginx template: %w", err)
		}

		conf = buf.String()
	}

	return conf, nil
}

func getPhpImage(phpVersion string) string {
	return fmt.Sprintf("php:%s-fpm", phpVersion)
}

func (p *PhpProvider) readComposerJson(ctx *generate.GenerateContext) (map[string]interface{}, error) {
	var composerJson map[string]interface{}
	err := ctx.App.ReadJSON("composer.json", &composerJson)
	if err != nil {
		return nil, err
	}

	return composerJson, nil
}
