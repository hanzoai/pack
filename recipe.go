package main

import (
	"github.com/moby/buildkit/client/llb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

// Base images, in one place. Go compiles to a static binary on distroless;
// Node and Python run on their official slim/alpine bases; static HTML is
// served by busybox httpd (no nginx/caddy).
const (
	goBuilder   = "docker.io/library/golang:1.23-alpine"
	goRuntime   = "gcr.io/distroless/static-debian12"
	nodeImage   = "docker.io/library/node:22-alpine"
	pythonImage = "docker.io/library/python:3.12-slim"
	staticImage = "docker.io/library/busybox:stable"
)

// Recipe is the concrete build for a Plan: the final rootfs graph plus the
// runtime config that turns it into a runnable image.
type Recipe struct {
	State      llb.State // final rootfs
	Base       string    // runtime image whose config is inherited
	Entrypoint []string
	WorkDir    string
	Ports      []string
}

// Recipe turns a Plan and the build context into a build. It is pure: the
// returned State marshals to a complete LLB graph without a running buildkitd.
func (p Plan) Recipe(src llb.State) Recipe {
	switch p.Ecosystem {
	case Go:
		built := llb.Image(goBuilder).
			With(copyDir(src, "/src")).
			Dir("/src").
			Run(llb.Shlex("go build -o /out/app ."), llb.AddEnv("CGO_ENABLED", "0")).Root()
		return Recipe{
			State:      llb.Image(goRuntime).With(copyFile(built, "/out/app", "/app")),
			Base:       goRuntime,
			Entrypoint: []string{"/app"},
		}

	case Node:
		install := "npm install"
		if p.Lockfile {
			install = "npm ci"
		}
		app := llb.Image(nodeImage).
			With(copyDir(src, "/app")).
			Dir("/app").
			Run(llb.Shlex(install)).Root().
			Run(llb.Shlex("npm run build --if-present")).Root()
		return Recipe{State: app, Base: nodeImage, Entrypoint: []string{"npm", "start"}, WorkDir: "/app"}

	case Python:
		install := "pip install --no-cache-dir ."
		if p.Requirements {
			install = "pip install --no-cache-dir -r requirements.txt"
		}
		app := llb.Image(pythonImage).
			With(copyDir(src, "/app")).
			Dir("/app").
			Run(llb.Shlex(install)).Root()
		return Recipe{State: app, Base: pythonImage, Entrypoint: []string{"python", "main.py"}, WorkDir: "/app"}

	case Static:
		return Recipe{
			State:      llb.Image(staticImage).With(copyDir(src, "/site")),
			Base:       staticImage,
			Entrypoint: []string{"httpd", "-f", "-v", "-p", "8080", "-h", "/site"},
			Ports:      []string{"8080/tcp"},
		}
	}
	panic("pack: Recipe called with unknown ecosystem " + string(p.Ecosystem))
}

// image overlays the recipe's runtime config onto the inherited base config,
// preserving the base env (PATH, so npm/python/httpd resolve) and never
// mutating the caller's base.
func (r Recipe) image(base *dockerspec.DockerOCIImage) *dockerspec.DockerOCIImage {
	img := *base
	img.Config.Entrypoint = r.Entrypoint
	img.Config.Cmd = nil
	if r.WorkDir != "" {
		img.Config.WorkingDir = r.WorkDir
	}
	img.Config.Env = append([]string{}, base.Config.Env...)
	if len(r.Ports) > 0 {
		ports := map[string]struct{}{}
		for k, v := range base.Config.ExposedPorts {
			ports[k] = v
		}
		for _, p := range r.Ports {
			ports[p] = struct{}{}
		}
		img.Config.ExposedPorts = ports
	}
	return &img
}

func copyDir(src llb.State, dst string) llb.StateOption {
	return func(s llb.State) llb.State {
		return s.File(llb.Copy(src, "/", dst, &llb.CopyInfo{CopyDirContentsOnly: true, CreateDestPath: true}))
	}
}

func copyFile(src llb.State, from, to string) llb.StateOption {
	return func(s llb.State) llb.State {
		return s.File(llb.Copy(src, from, to, &llb.CopyInfo{CreateDestPath: true}))
	}
}
