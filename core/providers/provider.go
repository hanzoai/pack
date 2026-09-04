package providers

import (
	"strings"

	"github.com/hanzoai/pack/core/generate"
	"github.com/hanzoai/pack/core/plan"
	"github.com/hanzoai/pack/core/providers/cpp"
	"github.com/hanzoai/pack/core/providers/deno"
	"github.com/hanzoai/pack/core/providers/dotnet"
	"github.com/hanzoai/pack/core/providers/elixir"
	"github.com/hanzoai/pack/core/providers/gleam"
	"github.com/hanzoai/pack/core/providers/golang"
	"github.com/hanzoai/pack/core/providers/java"
	"github.com/hanzoai/pack/core/providers/node"
	"github.com/hanzoai/pack/core/providers/php"
	"github.com/hanzoai/pack/core/providers/python"
	"github.com/hanzoai/pack/core/providers/ruby"
	"github.com/hanzoai/pack/core/providers/rust"
	"github.com/hanzoai/pack/core/providers/shell"
	"github.com/hanzoai/pack/core/providers/staticfile"
)

type Provider interface {
	Name() string
	Detect(ctx *generate.GenerateContext) (bool, error)
	Initialize(ctx *generate.GenerateContext) error
	Plan(ctx *generate.GenerateContext) error
	CleansePlan(buildPlan *plan.BuildPlan)
	StartCommandHelp() string
}

func GetLanguageProviders() []Provider {
	// Order is important here. The first provider that returns true from Detect() will be used.
	return []Provider{
		&php.PhpProvider{},
		&golang.GoProvider{},
		&java.JavaProvider{},
		&rust.RustProvider{},
		&ruby.RubyProvider{},
		&elixir.ElixirProvider{},
		&python.PythonProvider{},
		&deno.DenoProvider{},
		&dotnet.DotnetProvider{},
		&node.NodeProvider{},
		&gleam.GleamProvider{},
		&cpp.CppProvider{},
		&staticfile.StaticfileProvider{},
		&shell.ShellProvider{},
	}
}

func GetProvider(name string) Provider {
	for _, provider := range GetLanguageProviders() {
		if strings.EqualFold(provider.Name(), name) {
			return provider
		}
	}

	return nil
}
