package providers

import (
	"github.com/railwayapp/railpack-go/core/generate"
	"github.com/railwayapp/railpack-go/core/providers/golang"
	"github.com/railwayapp/railpack-go/core/providers/node"
	"github.com/railwayapp/railpack-go/core/providers/php"
	"github.com/railwayapp/railpack-go/core/providers/python"
)

type Provider interface {
	Name() string
	Plan(ctx *generate.GenerateContext) (bool, error)
}

func GetLanguageProviders() []Provider {
	return []Provider{
		&php.PhpProvider{},
		&node.NodeProvider{},
		&python.PythonProvider{},
		&golang.GoProvider{},
	}
}
