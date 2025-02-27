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
		outputDir string
	}{
		{
			name:      "npm",
			path:      "../../../examples/node-vite-react",
			isSPA:     true,
			outputDir: "dist",
		},
		{
			name:      "bun",
			path:      "../../../examples/node-vite-svelte",
			isSPA:     true,
			outputDir: "theoutput",
		},
		{
			name:      "pnpm",
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
			err := provider.Initialize(ctx)
			require.NoError(t, err)
			isSPA := provider.isSPA(ctx)
			require.Equal(t, tt.isSPA, isSPA)

			if tt.isSPA {
				require.Equal(t, tt.outputDir, provider.getOutputDirectory(ctx))
			}
		})
	}
}
