package python

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_PYTHON_VERSION = "3.13.1"
	UV_CACHE_DIR           = "/opt/uv-cache"
	PIP_CACHE_DIR          = "/opt/pip-cache"
)

type PythonProvider struct{}

func (p *PythonProvider) Name() string {
	return "python"
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
	if err := p.packages(ctx); err != nil {
		return err
	}

	if err := p.install(ctx); err != nil {
		return err
	}

	if err := p.start(ctx); err != nil {
		return err
	}

	p.addMetadata(ctx)

	return nil
}

func (p *PythonProvider) start(ctx *generate.GenerateContext) error {
	ctx.Start.AddOutputs([]string{"."})

	var startCommand string

	if ctx.App.HasMatch("main.py") {
		startCommand = "python main.py"
	}

	if startCommand != "" {
		ctx.Start.Command = startCommand
	}

	return nil
}

func (p *PythonProvider) install(ctx *generate.GenerateContext) error {

	hasRequirements := p.hasRequirements(ctx)
	hasPyproject := p.hasPyproject(ctx)
	hasPipfile := p.hasPipfile(ctx)
	hasPoetry := p.hasPoetry(ctx)
	hasPdm := p.hasPdm(ctx)
	hasUv := p.hasUv(ctx)

	setup := ctx.NewCommandStep("setup")
	setup.AddEnvVars(p.GetPythonEnvVars(ctx))
	setup.AddPaths([]string{"/root/.local/bin"})

	install := ctx.NewCommandStep("install")
	install.DependsOn = append(install.DependsOn, setup.DisplayName)

	if hasRequirements {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("requirements.txt"),
			plan.NewExecCommand("pip install -r requirements.txt", plan.ExecOptions{
				Caches: []string{ctx.Caches.AddCache("pip", PIP_CACHE_DIR)},
			}),
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
		// TODO: Fix this. PDM is not working because the packages are installed into a venv
		// that is not available to python at runtime
		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install pdm"),
			plan.NewVariableCommand("PDM_CHECK_UPDATE", "false"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("pdm.lock"),
			plan.NewCopyCommand("."),
			plan.NewExecCommand("pdm install --check --prod --no-editable"),
			plan.NewPathCommand("/app/.venv/bin"),
		})
	} else if hasPyproject && hasUv {
		install.AddCommands([]plan.Command{
			plan.NewVariableCommand("UV_COMPILE_BYTECODE", "1"),
			plan.NewVariableCommand("UV_LINK_MODE", "copy"),
			plan.NewVariableCommand("UV_CACHE_DIR", UV_CACHE_DIR),
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

	aptStep := ctx.NewAptStepBuilder("python-system-deps")
	aptStep.Packages = []string{"pkg-config"}
	install.DependsOn = append(install.DependsOn, aptStep.DisplayName)

	for dep, requiredPkgs := range pythonDepRequirements {
		if p.usesDep(ctx, dep) {
			aptStep.Packages = append(aptStep.Packages, requiredPkgs...)
		}
	}

	return nil
}

func (p *PythonProvider) packages(ctx *generate.GenerateContext) error {
	packages := ctx.GetMiseStepBuilder()

	python := packages.Default("python", DEFAULT_PYTHON_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("PYTHON_VERSION"); envVersion != "" {
		packages.Version(python, envVersion, varName)
	}

	if versionFile, err := ctx.App.ReadFile(".python-version"); err == nil {
		packages.Version(python, string(versionFile), ".python-version")
	}

	if runtimeFile, err := ctx.App.ReadFile("runtime.txt"); err == nil {
		packages.Version(python, string(runtimeFile), "runtime.txt")
	}

	if pipfileVersion := parseVersionFromPipfile(ctx); pipfileVersion != "" {
		packages.Version(python, pipfileVersion, "Pipfile")
	}

	if p.hasPoetry(ctx) || p.hasUv(ctx) || p.hasPdm(ctx) {
		packages.Default("pipx", "latest")
	}

	return nil
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
