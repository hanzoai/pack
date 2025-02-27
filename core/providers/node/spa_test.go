package node

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestVite(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isSPA     bool
		isVite    bool
		isAstro   bool
		outputDir string
	}{
		{
			name:      "vite-react",
			path:      "../../../examples/node-vite-react",
			isSPA:     true,
			isVite:    true,
			outputDir: "dist",
		},
		{
			name:      "vite-svelte",
			path:      "../../../examples/node-vite-svelte",
			isSPA:     true,
			isVite:    true,
			outputDir: "theoutput",
		},
		{
			name:      "astro-static",
			path:      "../../../examples/node-astro",
			isSPA:     true,
			isAstro:   true,
			outputDir: "dist",
		},
		{
			name:      "astro-server",
			path:      "../../../examples/node-astro-server",
			isSPA:     false,
			isAstro:   true,
			outputDir: "dist",
		},
		{
			name:      "corepack",
			path:      "../../../examples/node-corepack",
			isSPA:     false,
			outputDir: "",
		},
		{
			name:      "golang",
			path:      "../../../examples/go-mod",
			isSPA:     false,
			outputDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			if !detected {
				return
			}

			err = provider.Initialize(ctx)
			require.NoError(t, err)
			isSPA := provider.isSPA(ctx)
			require.Equal(t, tt.isSPA, isSPA)

			isVite := provider.isVite(ctx)
			require.Equal(t, tt.isVite, isVite)

			isAstro := provider.isAstro(ctx)
			require.Equal(t, tt.isAstro, isAstro)

			if tt.isSPA {
				require.Equal(t, tt.outputDir, provider.getOutputDirectory(ctx))
			}
		})
	}
}
