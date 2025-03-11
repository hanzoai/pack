package php

import (
	"fmt"
	"net/http"
	"strings"

	_ "embed"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/objx"
)

const (
	DEFAULT_PHP_VERSION  = "8.4.3"
	DefaultCaddyfilePath = "/Caddyfile"
)

//go:embed Caddyfile
var caddyfileTemplate string

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

	prepare := ctx.NewCommandStep("prepare")
	prepare.AddInput(plan.NewStepInput(phpImageStep.Name()))
	prepare.AddEnvVars(map[string]string{
		"SERVER_NAME": ":80",
		"APP_ENV":     "production",
		"APP_DEBUG":   "false",
		"PHP_INI_DIR": "/usr/local/etc/php",
	})
	prepare.AddCommands([]plan.Command{
		plan.NewExecCommand("echo 1"),
		plan.NewExecCommand("env"),
		plan.NewExecCommand("ls -la /usr/local/etc/php"),
		plan.NewExecShellCommand("cp $PHP_INI_DIR/php.ini-production $PHP_INI_DIR/php.ini"),
		plan.NewFileCommand(DefaultCaddyfilePath, "Caddyfile"),
	})

	configFiles, err := p.getConfigFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config files: %w", err)
	}
	prepare.Assets["Caddyfile"] = configFiles.Caddyfile

	// Composer
	composer := ctx.NewCommandStep("install:composer")
	composer.AddInput(plan.NewStepInput(prepare.Name()))
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

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepInput(composer.Name()))
	build.AddCommand(plan.NewCopyCommand("."))

	ctx.Deploy.StartCmd = fmt.Sprintf("docker-php-entrypoint --config %s --adapter caddyfile 2>&1", DefaultCaddyfilePath)

	ctx.Deploy.Inputs = []plan.Input{
		plan.NewStepInput(build.Name()),
	}

	return nil
}

// func (p *PhpProvider) Plan2(ctx *generate.GenerateContext) error {
// 	phpImageStep, err := p.phpImagePackage(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	prepare := ctx.NewCommandStep("prepare")
// 	prepare.AddInput(plan.NewStepInput(phpImageStep.Name()))
// 	prepare.AddEnvVars(map[string]string{
// 		"SERVER_NAME": ":80",
// 		"APP_ENV":     "production",
// 		"APP_DEBUG":   "false",
// 	})
// 	prepare.AddCommands([]plan.Command{
// 		plan.NewExecCommand("cp $PHP_INI_DIR/php.ini-production $PHP_INI_DIR/php.ini"),
// 	})

// 	// Nginx
// 	nginx := ctx.NewCommandStep("nginx")
// 	nginx.AddInput(plan.NewStepInput(phpImageStep.Name()))
// 	nginx.AddCommands([]plan.Command{
// 		plan.NewFileCommand("/etc/nginx/railpack.conf", "nginx.conf", plan.FileOptions{CustomName: "create nginx config"}),
// 		plan.NewExecCommand("nginx -t -c /etc/nginx/railpack.conf"),
// 		plan.NewFileCommand("/etc/php-fpm.conf", "php-fpm.conf", plan.FileOptions{CustomName: "create php-fpm config"}),
// 		plan.NewFileCommand("/start-nginx.sh", "start-nginx.sh", plan.FileOptions{
// 			CustomName: "create start nginx script",
// 			Mode:       0755,
// 		}),
// 	})

// 	nginx.Assets["start-nginx.sh"] = startNginxScriptAsset
// 	configFiles, err := p.getConfigFiles(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get config files: %w", err)
// 	}

// 	nginx.Assets["nginx.conf"] = configFiles.NginxConf
// 	nginx.Assets["php-fpm.conf"] = configFiles.PhpFpmConf

// 	// Composer
// 	composer := ctx.NewCommandStep("install:composer")
// 	composer.AddInput(plan.NewStepInput(prepare.Name()))
// 	if _, err := p.readComposerJson(ctx); err == nil {
// 		composer.AddCache(ctx.Caches.AddCache("composer", "/opt/cache/composer"))
// 		composer.AddEnvVars(map[string]string{"COMPOSER_CACHE_DIR": "/opt/cache/composer"})

// 		composer.AddCommands([]plan.Command{
// 			// Copy composer from the composer image
// 			plan.CopyCommand{Image: "composer:latest", Src: "/usr/bin/composer", Dest: "/usr/bin/composer"},
// 			plan.NewCopyCommand("."),
// 			plan.NewExecCommand("composer install --ignore-platform-reqs"),
// 		})
// 	}

// 	// Node (if necessary)
// 	nodeProvider := node.NodeProvider{}
// 	isNode, err := nodeProvider.Detect(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	if p.usesLaravel(ctx) {
// 		ctx.Logger.LogInfo("Found Laravel app")
// 	}

// 	if isNode {
// 		err = nodeProvider.Initialize(ctx)
// 		if err != nil {
// 			return err
// 		}

// 		ctx.Logger.LogInfo("Installing Node")

// 		miseStep := ctx.GetMiseStepBuilder()
// 		nodeProvider.InstallMisePackages(ctx, miseStep)

// 		install := ctx.NewCommandStep("install:node")
// 		install.AddInput(plan.NewStepInput(miseStep.Name()))
// 		nodeProvider.InstallNodeDeps(ctx, install)

// 		prune := ctx.NewCommandStep("prune:node")
// 		prune.AddInput(plan.NewStepInput(install.Name()))
// 		nodeProvider.PruneNodeDeps(ctx, prune)

// 		build := ctx.NewCommandStep("build")
// 		build.Inputs = []plan.Input{
// 			plan.NewStepInput(composer.Name()),

// 			// Include the app and mise packages (node, pnpm, etc.)
// 			plan.NewStepInput(install.Name(), plan.InputOptions{
// 				Include: append([]string{"."}, miseStep.GetOutputPaths()...),
// 			}),
// 		}
// 		nodeProvider.Build(ctx, build)

// 		ctx.Deploy.Inputs = []plan.Input{
// 			plan.NewStepInput(composer.Name()),
// 			plan.NewStepInput(build.Name(), plan.InputOptions{
// 				Include: []string{"."},
// 				Exclude: []string{"node_modules"},
// 			}),
// 			plan.NewStepInput(prune.Name(), plan.InputOptions{
// 				Include: []string{"/app/node_modules"},
// 			}),
// 			plan.NewLocalInput("."),
// 		}
// 	} else {
// 		// A manual build command will go here
// 		build := ctx.NewCommandStep("build")
// 		build.AddInput(plan.NewStepInput(composer.Name()))

// 		fmt.Printf("build: %v\n", build)

// 		ctx.Deploy.Inputs = []plan.Input{
// 			plan.NewStepInput(build.Name()),
// 			plan.NewLocalInput("."),
// 		}
// 	}

// 	ctx.Deploy.StartCmd = "bash /start-nginx.sh"

// 	if p.usesLaravel(ctx) {
// 		ctx.Deploy.Variables["IS_LARAVEL"] = "true"
// 	}

// 	return nil
// }

func (p *PhpProvider) usesLaravel(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("artisan")
}

type ConfigFiles struct {
	NginxConf  string
	PhpFpmConf string
	Caddyfile  string
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

	fmt.Printf("data: %v\n", data)

	caddyfile, err := ctx.TemplateFiles([]string{"Caddyfile"}, caddyfileTemplate, data)
	if err != nil {
		return nil, err
	}

	fmt.Printf("caddyfile: %v\n", caddyfile)

	return &ConfigFiles{
		Caddyfile: caddyfile.Contents,
	}, nil
}

func (p *PhpProvider) getConfigFiles2(ctx *generate.GenerateContext) (*ConfigFiles, error) {
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

	nginxConfTemplate, err := ctx.TemplateFiles([]string{"nginx.template.conf", "nginx.conf"}, nginxConfTemplateAsset, data)
	if err != nil {
		return nil, err
	}

	phpFpmConf, err := ctx.TemplateFiles([]string{"php-fpm.template.conf", "php-fpm.conf"}, phpFpmConfTemplateAsset, data)
	if err != nil {
		return nil, err
	}

	if nginxConfTemplate.Filename != "" {
		ctx.Logger.LogInfo("Using custom nginx config: %s", nginxConfTemplate.Filename)
	}

	if phpFpmConf.Filename != "" {
		ctx.Logger.LogInfo("Using custom php-fpm config: %s", phpFpmConf.Filename)
	}

	return &ConfigFiles{
		NginxConf:  nginxConfTemplate.Contents,
		PhpFpmConf: phpFpmConf.Contents,
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

	// imageStep.AptPackages = append(imageStep.AptPackages, "nginx", "git", "zip", "unzip")

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

	// Ensure that the version is available on Docker Hub
	imageStep.SetVersionAvailable(php, func(version string) bool {
		image := getPhpImage(version)
		fmt.Println("Checking if image is available:", image)

		// dunglas/frankenphp:php8.4.3-bookworm -> [dunglas, frankenphp, php8.4.3-bookworm]
		parts := strings.Split(image, ":")
		repository := parts[0] // dunglas/frankenphp
		tag := parts[1]        // php8.4.3-bookworm

		url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/%s", repository, tag)
		fmt.Println("URL:", url)
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	})

	return imageStep, nil
}

func getPhpImage(phpVersion string) string {
	return fmt.Sprintf("dunglas/frankenphp:php%s-bookworm", phpVersion)
}

func (p *PhpProvider) readComposerJson(ctx *generate.GenerateContext) (map[string]interface{}, error) {
	var composerJson map[string]interface{}
	err := ctx.App.ReadJSON("composer.json", &composerJson)
	if err != nil {
		return nil, err
	}

	return composerJson, nil
}

func (p *PhpProvider) StartCommandHelp() string {
	return ""
}
