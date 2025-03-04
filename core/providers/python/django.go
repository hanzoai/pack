package python

import (
	"fmt"
	"regexp"

	"github.com/railwayapp/railpack/core/generate"
)

func (p *PythonProvider) getDjangoAppName(ctx *generate.GenerateContext) (string, error) {
	paths, err := ctx.App.FindFiles("**/*.py")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`WSGI_APPLICATION = ["'](.*).application["']`)

	for _, path := range paths {
		contents, err := ctx.App.ReadFile(path)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(contents)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("failed to find your WSGI_APPLICATION django setting")
}

func (p *PythonProvider) isDjango(ctx *generate.GenerateContext) bool {
	hasManage := ctx.App.HasMatch("manage.py")
	importsDjango := p.usesDep(ctx, "django")

	return hasManage && importsDjango
}
