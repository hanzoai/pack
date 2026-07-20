package main

import (
	"github.com/moby/buildkit/client/llb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

// Base images, in one place. Go compiles to a static binary on distroless;
// Node and Python run on their official slim/alpine bases; a site is files, so
// it has no base at all (no webserver, no nginx/caddy).
const (
	goBuilder   = "docker.io/library/golang:1.23-alpine"
	goRuntime   = "gcr.io/distroless/static-debian12"
	nodeImage   = "docker.io/library/node:22-alpine"
	pythonImage = "docker.io/library/python:3.12-slim"
)

// siteOutput stages the built site at /out from the first output directory
// that exists, in this order. cp -a keeps symlinks as symlinks, so nothing
// outside the output directory is ever pulled in. No output directory is a
// hard failure: exporting the context instead would publish the source.
const siteOutput = `mkdir -p /out; for d in dist build out public _site; do ` +
	`if [ -d "$d" ]; then cp -a "$d/." /out/; exit 0; fi; done; ` +
	`echo 'pack: no build output (looked for dist build out public _site)' >&2; exit 1`

// Recipe is the concrete build for a Plan. A recipe with a Base is an image:
// the final rootfs plus the runtime config that makes it runnable. A recipe
// without one is files — a built site, which needs no runtime.
type Recipe struct {
	State      llb.State // final rootfs, or the site's files
	Base       string    // runtime image whose config is inherited; empty for files
	Entrypoint []string
	WorkDir    string
	Ports      []string
}

// files reports whether the recipe is a filesystem result rather than an
// image. A site is served by the ingress staticFiles plugin from object
// storage, so it leaves as the files it is: no image, no pod, no compute.
func (r Recipe) files() bool { return r.Base == "" }

// Recipe turns a Plan and the build context into a build. It is pure: the
// returned State marshals to a complete LLB graph without a running buildkitd.
func (p Plan) Recipe(src llb.State) Recipe {
	switch p.Ecosystem {
	case Go:
		built := llb.Image(goBuilder).
			With(copyDir(src, "/", "/src")).
			Dir("/src").
			Run(llb.Shlex("go build -o /out/app ."), llb.AddEnv("CGO_ENABLED", "0")).Root()
		return Recipe{
			State:      llb.Image(goRuntime).With(copyFile(built, "/out/app", "/app")),
			Base:       goRuntime,
			Entrypoint: []string{"/app"},
		}

	case Node:
		app := llb.Image(nodeImage).
			With(copyDir(src, "/", "/app")).
			Dir("/app").
			Run(llb.Shlex(npmInstall(p.Lockfile))).Root().
			Run(llb.Shlex("npm run build --if-present")).Root()
		return Recipe{State: app, Base: nodeImage, Entrypoint: []string{"npm", "start"}, WorkDir: "/app"}

	case Python:
		install := "pip install --no-cache-dir ."
		if p.Requirements {
			install = "pip install --no-cache-dir -r requirements.txt"
		}
		app := llb.Image(pythonImage).
			With(copyDir(src, "/", "/app")).
			Dir("/app").
			Run(llb.Shlex(install)).Root()
		return Recipe{State: app, Base: pythonImage, Entrypoint: []string{"python", "main.py"}, WorkDir: "/app"}

	case Static:
		// Already files: export the context as it is.
		if !p.Generated {
			return Recipe{State: src}
		}
		// A generator: run its build, then export only what the build wrote.
		built := llb.Image(nodeImage).
			With(copyDir(src, "/", "/site")).
			Dir("/site").
			Run(llb.Shlex(npmInstall(p.Lockfile))).Root().
			Run(llb.Shlex("npm run build")).Root().
			Run(llb.Args([]string{"sh", "-c", siteOutput})).Root()
		return Recipe{State: llb.Scratch().With(copyDir(built, "/out", "/"))}
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

// npmInstall pins the tree with the lockfile when there is one. One fact, one
// place: the install is the same in the image lane and the site lane.
func npmInstall(lockfile bool) string {
	if lockfile {
		return "npm ci"
	}
	return "npm install"
}

func copyDir(src llb.State, from, to string) llb.StateOption {
	return func(s llb.State) llb.State {
		return s.File(llb.Copy(src, from, to, &llb.CopyInfo{CopyDirContentsOnly: true, CreateDestPath: true}))
	}
}

func copyFile(src llb.State, from, to string) llb.StateOption {
	return func(s llb.State) llb.State {
		return s.File(llb.Copy(src, from, to, &llb.CopyInfo{CreateDestPath: true}))
	}
}
