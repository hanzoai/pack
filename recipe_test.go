package main

import (
	"context"
	"strings"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

// graph marshals a State and returns every source image identifier and every
// exec argv in the LLB DAG, so recipes can be asserted without a buildkitd.
func graph(t *testing.T, st llb.State) (sources []string, execs [][]string) {
	t.Helper()
	def, err := st.Marshal(context.Background())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, dt := range def.Def {
		var op pb.Op
		if err := op.Unmarshal(dt); err != nil {
			t.Fatalf("unmarshal op: %v", err)
		}
		if s := op.GetSource(); s != nil {
			sources = append(sources, s.Identifier)
		}
		if e := op.GetExec(); e != nil && e.Meta != nil {
			execs = append(execs, e.Meta.Args)
		}
	}
	return sources, execs
}

func hasSource(sources []string, sub string) bool {
	for _, s := range sources {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func hasExec(execs [][]string, sub string) bool {
	for _, a := range execs {
		if strings.Contains(strings.Join(a, " "), sub) {
			return true
		}
	}
	return false
}

func recipeFor(t *testing.T, files ...string) Recipe {
	t.Helper()
	plan, err := Detect(fileset(files...))
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	return plan.Recipe(llb.Local("context"))
}

func TestGoRecipe(t *testing.T) {
	r := recipeFor(t, "go.mod")
	sources, execs := graph(t, r.State)
	if !hasSource(sources, "golang:1.23-alpine") {
		t.Errorf("missing go builder image; sources=%v", sources)
	}
	if !hasSource(sources, "distroless/static") {
		t.Errorf("missing distroless runtime; sources=%v", sources)
	}
	if !hasExec(execs, "go build -o /out/app .") {
		t.Errorf("missing go build step; execs=%v", execs)
	}
	if len(r.Entrypoint) != 1 || r.Entrypoint[0] != "/app" {
		t.Errorf("entrypoint = %v", r.Entrypoint)
	}
}

func TestNodeRecipe(t *testing.T) {
	r := recipeFor(t, "package.json", "package-lock.json")
	sources, execs := graph(t, r.State)
	if !hasSource(sources, "node:22-alpine") {
		t.Errorf("missing node image; sources=%v", sources)
	}
	if !hasExec(execs, "npm ci") {
		t.Errorf("lockfile should build with npm ci; execs=%v", execs)
	}
	if !hasExec(execs, "npm run build --if-present") {
		t.Errorf("missing npm build step; execs=%v", execs)
	}
	if got := strings.Join(r.Entrypoint, " "); got != "npm start" {
		t.Errorf("entrypoint = %q", got)
	}

	// Without a lockfile the install is npm install.
	_, execs = graph(t, recipeFor(t, "package.json").State)
	if !hasExec(execs, "npm install") {
		t.Errorf("no lockfile should build with npm install; execs=%v", execs)
	}
}

func TestPythonRecipe(t *testing.T) {
	r := recipeFor(t, "requirements.txt")
	sources, execs := graph(t, r.State)
	if !hasSource(sources, "python:3.12-slim") {
		t.Errorf("missing python image; sources=%v", sources)
	}
	if !hasExec(execs, "pip install --no-cache-dir -r requirements.txt") {
		t.Errorf("missing pip install; execs=%v", execs)
	}
	if got := strings.Join(r.Entrypoint, " "); got != "python main.py" {
		t.Errorf("entrypoint = %q", got)
	}

	// pyproject-only installs the project itself.
	_, execs = graph(t, recipeFor(t, "pyproject.toml").State)
	if !hasExec(execs, "pip install --no-cache-dir .") {
		t.Errorf("pyproject should pip install .; execs=%v", execs)
	}
}

func TestStaticRecipe(t *testing.T) {
	r := recipeFor(t, "index.html")
	sources, _ := graph(t, r.State)
	if !hasSource(sources, "busybox") {
		t.Errorf("missing busybox image; sources=%v", sources)
	}
	if got := strings.Join(r.Entrypoint, " "); got != "httpd -f -v -p 8080 -h /site" {
		t.Errorf("entrypoint = %q", got)
	}
	if len(r.Ports) != 1 || r.Ports[0] != "8080/tcp" {
		t.Errorf("ports = %v", r.Ports)
	}
}

func TestImageOverlay(t *testing.T) {
	base := &dockerspec.DockerOCIImage{}
	base.Config.Env = []string{"PATH=/usr/local/bin:/usr/bin"}
	base.Config.Cmd = []string{"node"}

	img := recipeFor(t, "package.json").image(base)
	if strings.Join(img.Config.Entrypoint, " ") != "npm start" {
		t.Errorf("entrypoint = %v", img.Config.Entrypoint)
	}
	if img.Config.WorkingDir != "/app" {
		t.Errorf("workdir = %q", img.Config.WorkingDir)
	}
	if len(img.Config.Cmd) != 0 {
		t.Errorf("base cmd should be cleared, got %v", img.Config.Cmd)
	}
	if len(img.Config.Env) != 1 || img.Config.Env[0] != "PATH=/usr/local/bin:/usr/bin" {
		t.Errorf("base PATH must be preserved, got %v", img.Config.Env)
	}
	// Overlay must not mutate the caller's base config.
	if base.Config.WorkingDir != "" || len(base.Config.Cmd) != 1 {
		t.Errorf("base config was mutated: %+v", base.Config)
	}

	static := recipeFor(t, "index.html").image(&dockerspec.DockerOCIImage{})
	if _, ok := static.Config.ExposedPorts["8080/tcp"]; !ok {
		t.Errorf("static image should expose 8080/tcp, got %v", static.Config.ExposedPorts)
	}
}
