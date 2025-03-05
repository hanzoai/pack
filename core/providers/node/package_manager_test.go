package node

import (
	"testing"
)

func TestParsePackageManagerField(t *testing.T) {
	tests := []struct {
		name           string
		packageManager string
		wantName       string
		wantVersion    string
	}{
		{
			name:           "valid package manager field",
			packageManager: "pnpm@8.15.4",
			wantName:       "pnpm",
			wantVersion:    "8.15.4",
		},
		{
			name:           "valid package manager field with yarn",
			packageManager: "yarn@4.1.0",
			wantName:       "yarn",
			wantVersion:    "4.1.0",
		},
		{
			name:           "valid package manager field with bun",
			packageManager: "bun@1.0.25",
			wantName:       "bun",
			wantVersion:    "1.0.25",
		},
		{
			name:           "valid package manager field with yarn and SHA",
			packageManager: "yarn@3.2.3+sha224.953c8233f7a92884eee2de69a1b92d1f2ec1655e66d08071ba9a02fa",
			wantName:       "yarn",
			wantVersion:    "3.2.3",
		},
		{
			name:           "empty package manager field",
			packageManager: "",
			wantName:       "",
			wantVersion:    "",
		},
		{
			name:           "invalid format - no version",
			packageManager: "pnpm",
			wantName:       "",
			wantVersion:    "",
		},
		{
			name:           "invalid format - multiple @ symbols",
			packageManager: "pnpm@8.15.4@extra",
			wantName:       "",
			wantVersion:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pm *string
			if tt.packageManager != "" {
				pm = &tt.packageManager
			}
			pkgJson := &PackageJson{
				PackageManager: pm,
			}
			gotName, gotVersion := PackageManager("").parsePackageManagerField(pkgJson)
			if gotName != tt.wantName || gotVersion != tt.wantVersion {
				t.Errorf("parsePackageManagerField() = (%v, %v), want (%v, %v)",
					gotName, gotVersion, tt.wantName, tt.wantVersion)
			}
		})
	}
}
