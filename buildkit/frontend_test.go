package buildkit

import (
	"testing"
)

func TestParseBuildArgs(t *testing.T) {
	opts := map[string]string{
		"build-arg:FOO": "bar",
		"platform":      "linux/amd64",
		"build-arg:BAZ": "qux",
		"filename":      "Dockerfile",
	}

	got := parseBuildArgs(opts)

	want := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	if len(got) != len(want) {
		t.Errorf("got %d build args, want %d", len(got), len(want))
	}

	for k, v := range want {
		if got[k] != v {
			t.Errorf("build arg %q = %q, want %q", k, got[k], v)
		}
	}
}
