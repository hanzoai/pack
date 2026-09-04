package cpp

import (
	"github.com/hanzoai/pack/core/generate"
	"github.com/hanzoai/pack/core/plan"
)

type cmake struct{}

func (p *CppProvider) DetectCmake(ctx *generate.GenerateContext) (buildSystem, bool) {
	if ctx.App.HasFile("CMakeLists.txt") {
		return &cmake{}, true
	}
	return nil, false
}

func (c *cmake) Install(ctx *generate.GenerateContext, mise *generate.MiseStepBuilder) {
	mise.Default("cmake", "latest")
	mise.Default("ninja", "latest")
	mise.UseMiseVersions(ctx, []string{"meson", "ninja"})
}

func (c *cmake) Build(build *generate.CommandStepBuilder) {
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("cmake -B /build -GNinja /app"),
		plan.NewExecCommand("cmake --build /build"),
	})
}
