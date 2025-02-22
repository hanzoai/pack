package python

import (
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_PYTHON_VERSION = "3.13.2"
	UV_CACHE_DIR           = "/opt/uv-cache"
	PIP_CACHE_DIR          = "/opt/pip-cache"
	PACKAGES_DIR           = "/opt/python-packages"
)

type PythonProvider struct{}

func (p *PythonProvider) Name() string {
	return "python"
}

func (p *PythonProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *PythonProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasPython := ctx.App.HasMatch("main.py") ||
		p.hasRequirements(ctx) ||
		p.hasPyproject(ctx) ||
		p.hasPoetry(ctx) ||
		p.hasPdm(ctx)

	return hasPython, nil
}

func (p *PythonProvider) Plan(ctx *generate.GenerateContext) error {
	// Install mise packages
	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	if p.hasRequirements(ctx) {
		p.PlanPip(ctx)
	}

	// Install dependencies
	// install := ctx.NewCommandStep("install")
	// install.AddInput(plan.NewStepInput(miseStep.Name()))
	// p.InstallPythonDeps(ctx, install)

	// // Build step (if needed)
	// build := ctx.NewCommandStep("build")
	// build.AddInput(plan.NewStepInput(install.Name()))

	// // Deploy configuration
	// ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	// maps.Copy(ctx.Deploy.Variables, p.GetPythonEnvVars(ctx))

	// ctx.Deploy.Inputs = append(ctx.Deploy.Inputs, []plan.Input{
	// 	plan.NewImageInput(plan.RAILPACK_RUNTIME_IMAGE),
	// 	plan.NewStepInput(miseStep.Name(), plan.InputOptions{
	// 		Include: miseStep.GetOutputPaths(),
	// 	}),
	// 	plan.NewStepInput(build.Name(), plan.InputOptions{
	// 		Include: []string{"."},
	// 	}),
	// 	plan.NewLocalInput("."),
	// }...)

	p.addMetadata(ctx)

	return nil
}

func (p *PythonProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	if ctx.App.HasMatch("main.py") {
		return "python main.py"
	}
	return ""
}

func (p *PythonProvider) PlanPip(ctx *generate.GenerateContext) {
	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepInput(ctx.GetMiseStepBuilder().Name()))

	install.AddCache(ctx.Caches.AddCache("pip", PIP_CACHE_DIR))
	install.AddCommands([]plan.Command{
		plan.NewCopyCommand("requirements.txt"),
		plan.NewExecCommand(fmt.Sprintf("pip install --target=%s -r requirements.txt", PACKAGES_DIR)),
	})
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"PYTHON", "PIP", "PIPX"})
	maps.Copy(install.Variables, p.GetPythonEnvVars(ctx))
	maps.Copy(install.Variables, map[string]string{
		"PIP_CACHE_DIR": PIP_CACHE_DIR,
		"PYTHONPATH":    PACKAGES_DIR,
	})

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepInput(install.Name()))

	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	maps.Copy(ctx.Deploy.Variables, p.GetPythonEnvVars(ctx))

	ctx.Deploy.Inputs = []plan.Input{
		plan.NewImageInput(plan.RAILPACK_RUNTIME_IMAGE),
		plan.NewStepInput(ctx.GetMiseStepBuilder().Name(), plan.InputOptions{
			Include: ctx.GetMiseStepBuilder().GetOutputPaths(),
		}),
		plan.NewStepInput(build.Name(), plan.InputOptions{
			Include: []string{PACKAGES_DIR, "."},
		}),
		plan.NewLocalInput("."),
	}
}

func (p *PythonProvider) InstallPythonDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	maps.Copy(install.Variables, p.GetPythonEnvVars(ctx))
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"PYTHON", "PIP", "PIPX", "PIPENV", "UV", "POETRY", "PDM"})
	install.AddPaths([]string{"/root/.local/bin", "/app/.venv/bin"})

	hasRequirements := p.hasRequirements(ctx)
	hasPyproject := p.hasPyproject(ctx)
	hasPipfile := p.hasPipfile(ctx)
	hasPoetry := p.hasPoetry(ctx)
	hasPdm := p.hasPdm(ctx)
	hasUv := p.hasUv(ctx)

	if hasRequirements {
		install.AddCache(ctx.Caches.AddCache("pip", PIP_CACHE_DIR))
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("requirements.txt"),
			plan.NewExecCommand("pip install -r requirements.txt"),
		})
	} else if hasPyproject && hasPoetry {
		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install poetry"),
			plan.NewExecCommand("poetry config virtualenvs.create false"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("poetry.lock"),
			plan.NewExecCommand("poetry install --no-interaction --no-ansi --no-root"),
		})
	} else if hasPyproject && hasPdm {
		install.AddEnvVars(map[string]string{"PDM_CHECK_UPDATE": "false"})
		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install pdm"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("pdm.lock"),
			plan.NewCopyCommand("."),
			plan.NewExecCommand("pdm install --check --prod --no-editable"),
			plan.NewPathCommand("/app/.venv/bin"),
		})
	} else if hasPyproject && hasUv {
		install.AddEnvVars(map[string]string{
			"UV_COMPILE_BYTECODE": "1",
			"UV_LINK_MODE":        "copy",
			"UV_CACHE_DIR":        UV_CACHE_DIR,
		})

		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install uv"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("uv.lock"),
			plan.NewExecCommand("uv sync --frozen --no-install-project --no-install-workspace --no-dev"),
			plan.NewCopyCommand("."),
			plan.NewExecCommand("uv sync --frozen --no-dev"),
			plan.NewPathCommand("/app/.venv/bin"),
		})
	} else if hasPipfile {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("Pipfile"),
		})

		if ctx.App.HasMatch("Pipfile.lock") {
			install.AddCommands([]plan.Command{
				plan.NewCopyCommand("Pipfile.lock"),
				plan.NewExecCommand("pipenv install --deploy"),
			})
		} else {
			install.AddCommands([]plan.Command{
				plan.NewExecCommand("pipenv install --skip-lock"),
			})
		}
	}

	// Handle system dependencies
	aptStep := ctx.NewAptStepBuilder("python-system-deps")
	aptStep.Packages = []string{"pkg-config"}

	for dep, requiredPkgs := range pythonDepRequirements {
		if p.usesDep(ctx, dep) {
			aptStep.Packages = append(aptStep.Packages, requiredPkgs...)
		}
	}
}

func (p *PythonProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	python := miseStep.Default("python", DEFAULT_PYTHON_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("PYTHON_VERSION"); envVersion != "" {
		miseStep.Version(python, envVersion, varName)
	}

	if versionFile, err := ctx.App.ReadFile(".python-version"); err == nil {
		miseStep.Version(python, string(versionFile), ".python-version")
	}

	if runtimeFile, err := ctx.App.ReadFile("runtime.txt"); err == nil {
		miseStep.Version(python, string(runtimeFile), "runtime.txt")
	}

	if pipfileVersion := parseVersionFromPipfile(ctx); pipfileVersion != "" {
		miseStep.Version(python, pipfileVersion, "Pipfile")
	}

	if p.hasPoetry(ctx) || p.hasUv(ctx) || p.hasPdm(ctx) {
		miseStep.Default("pipx", "latest")
	}
}

func (p *PythonProvider) GetPythonEnvVars(ctx *generate.GenerateContext) map[string]string {
	return map[string]string{
		"PYTHONFAULTHANDLER":            "1",
		"PYTHONUNBUFFERED":              "1",
		"PYTHONHASHSEED":                "random",
		"PYTHONDONTWRITEBYTECODE":       "1",
		"PIP_DISABLE_PIP_VERSION_CHECK": "1",
		"PIP_DEFAULT_TIMEOUT":           "100",
		"PIP_CACHE_DIR":                 PIP_CACHE_DIR,
	}
}

func (p *PythonProvider) addMetadata(ctx *generate.GenerateContext) {
	hasPoetry := p.hasPoetry(ctx)
	hasPdm := p.hasPdm(ctx)
	hasUv := p.hasUv(ctx)

	pkgManager := "pip"

	if hasPoetry {
		pkgManager = "poetry"
	} else if hasPdm {
		pkgManager = "pdm"
	} else if hasUv {
		pkgManager = "uv"
	}

	ctx.Metadata.Set("packageManager", pkgManager)
	ctx.Metadata.SetBool("requirements", p.hasRequirements(ctx))
	ctx.Metadata.SetBool("pyproject", p.hasPyproject(ctx))
	ctx.Metadata.SetBool("pipfile", p.hasPipfile(ctx))
}

func (p *PythonProvider) usesDep(ctx *generate.GenerateContext, dep string) bool {
	for _, file := range []string{"requirements.txt", "pyproject.toml", "Pipfile"} {
		if contents, err := ctx.App.ReadFile(file); err == nil {
			if strings.Contains(strings.ToLower(contents), strings.ToLower(dep)) {
				return true
			}
		}
	}
	return false
}

var pipfileVersionRegex = regexp.MustCompile(`(python_version|python_full_version)\s*=\s*['"]([0-9.]*)"?`)

func parseVersionFromPipfile(ctx *generate.GenerateContext) string {
	pipfile, err := ctx.App.ReadFile("Pipfile")
	if err != nil {
		return ""
	}

	matches := pipfileVersionRegex.FindStringSubmatch(string(pipfile))

	if len(matches) > 2 {
		return matches[2]
	}
	return ""
}

func (p *PythonProvider) hasRequirements(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("requirements.txt")
}

func (p *PythonProvider) hasPyproject(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("pyproject.toml")
}

func (p *PythonProvider) hasPipfile(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("Pipfile")
}

func (p *PythonProvider) hasPoetry(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("poetry.lock")
}

func (p *PythonProvider) hasPdm(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("pdm.lock")
}

func (p *PythonProvider) hasUv(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("uv.lock")
}

// Mapping of python dependencies to required apt packages
var pythonDepRequirements = map[string][]string{
	"pdf2image": {"poppler-utils", "gcc"},
	"pydub":     {"ffmpeg", "gcc"},
	"pymovie":   {"ffmpeg", "qt5-qmake", "qtbase5-dev", "qtbase5-dev-tools", "qttools5-dev-tools", "libqt5core5a", "python3-pyqt5", "gcc"},
}
