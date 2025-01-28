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
	DEFAULT_PHP_VERSION = "8.4"
)

type PhpProvider struct{}

func (p *PhpProvider) Name() string {
	return "php"
}

func (p *PhpProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("index.php") ||
		ctx.App.HasMatch("composer.json"), nil
}

func (p *PhpProvider) Plan(ctx *generate.GenerateContext) error {
	imageStep, err := p.phpImagePackage(ctx)
	if err != nil {
		return err
	}

	// Install nginx
	nginxPackages := ctx.NewAptStepBuilder("nginx")
	nginxPackages.Packages = []string{"nginx", "git", "zip", "unzip"}
	nginxPackages.DependsOn = []string{imageStep.DisplayName}

	// Install composer
	if _, err := p.readComposerJson(ctx); err == nil {
		install := ctx.NewCommandStep("install")
		install.AddCommands([]plan.Command{
			// Copy composer from the composer image
			plan.CopyCommand{Image: "composer:latest", Src: "/usr/bin/composer", Dest: "/usr/bin/composer"},
			plan.NewCopyCommand("."),
			plan.NewExecCommand("composer install --ignore-platform-reqs"),
		})

		install.DependsOn = []string{nginxPackages.DisplayName}
	}

	// Install node
	nodeProvider := node.NodeProvider{}
	if packageJson, err := nodeProvider.GetPackageJson(ctx.App); err == nil && packageJson != nil {
		ctx.EnterSubContext("node")

		nodePackages, err := nodeProvider.Packages(ctx, packageJson)
		if err != nil {
			return err
		}
		nodePackages.DependsOn = []string{imageStep.DisplayName}

		nodeInstall, err := nodeProvider.Install(ctx, nodePackages, packageJson)
		if err != nil {
			return err
		}

		_, err = nodeProvider.Build(ctx, nodeInstall, packageJson)
		if err != nil {
			return err
		}

		ctx.ExitSubContext()
	}

	// Setup nginx
	nginxSetup := ctx.NewCommandStep("nginx:setup")
	nginxSetup.DependsOn = []string{nginxPackages.DisplayName}

	nginxSetup.AddCommands([]plan.Command{
		plan.NewFileCommand("/etc/nginx/railpack.conf", "nginx.conf", plan.FileOptions{CustomName: "create nginx config"}),
		plan.NewExecCommand("nginx -t -c /etc/nginx/railpack.conf"),
		plan.NewFileCommand("/etc/php-fpm.conf", "php-fpm.conf", plan.FileOptions{CustomName: "create php-fpm config"}),
		plan.NewFileCommand("/start-nginx.sh", "start-nginx.sh", plan.FileOptions{
			CustomName: "create start nginx script",
			Mode:       0755,
		}),
	})

	if p.usesLaravel(ctx) {
		nginxSetup.AddCommands([]plan.Command{
			plan.NewVariableCommand("IS_LARAVEL", "true"),
		})
	}

	nginxSetup.Assets["start-nginx.sh"] = startNginxScriptAsset
	configFiles, err := p.getConfigFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config files: %w", err)
	}

	nginxSetup.Assets["nginx.conf"] = configFiles.NginxConf
	nginxSetup.Assets["nginx.conf"] = configFiles.NginxConf
	nginxSetup.Assets["php-fpm.conf"] = configFiles.PhpFpmConf

	ctx.Start.Command = "bash /start-nginx.sh"
	ctx.Start.Paths = []string{"."}

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
	imageStep := ctx.NewImageStep("packages", func(options *generate.BuildStepOptions) string {
		if phpVersion, ok := options.ResolvedPackages["php"]; ok && phpVersion.ResolvedVersion != nil {
			return getPhpImage(*phpVersion.ResolvedVersion)
		}

		// Return the default if we were not able to resolve the version
		return getPhpImage(DEFAULT_PHP_VERSION)
	})

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
