package node

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "npm",
			path: "../../../examples/node-npm-latest",
			want: true,
		},
		{
			name: "bun",
			path: "../../../examples/node-bun",
			want: true,
		},
		{
			name: "pnpm",
			path: "../../../examples/node-corepack",
			want: true,
		},
		{
			name: "golang",
			path: "../../../examples/go-mod",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPackageManager(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		packageManager PackageManager
	}{
		{
			name:           "npm project",
			path:           "../../../examples/node-npm-latest",
			packageManager: PackageManagerNpm,
		},
		{
			name:           "bun project",
			path:           "../../../examples/node-bun",
			packageManager: PackageManagerBun,
		},
		{
			name:           "pnpm project",
			path:           "../../../examples/node-corepack",
			packageManager: PackageManagerPnpm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			packageManager := provider.getPackageManager(ctx.App)
			require.Equal(t, tt.packageManager, packageManager)
		})
	}
}

func TestNodeCorepack(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantCorepack bool
	}{
		{
			name:         "corepack project",
			path:         "../../../examples/node-corepack",
			wantCorepack: true,
		},
		{
			name:         "bun project",
			path:         "../../../examples/node-bun",
			wantCorepack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			packageJson, err := provider.GetPackageJson(ctx.App)
			require.NoError(t, err)

			usesCorepack := provider.usesCorepack(packageJson)
			require.Equal(t, tt.wantCorepack, usesCorepack)
		})
	}
}
