package main

import (
	"context"
	"encoding/json"
	"path"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/frontend/gateway/client"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

// maxPackage bounds the package.json read at the context boundary.
const maxPackage = 1 << 20

// build is the gateway callback. It loads the build context, detects the
// ecosystem once, then builds each requested platform through dockerui so the
// result carries the image config the exporter needs.
func build(ctx context.Context, c client.Client) (*client.Result, error) {
	bc, err := dockerui.NewClient(c)
	if err != nil {
		return nil, err
	}
	src, err := bc.MainContext(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := detectContext(ctx, c, *src)
	if err != nil {
		return nil, err
	}
	// A site leaves as files; everything else leaves as an image.
	if recipe := plan.Recipe(*src); recipe.files() {
		return solveFiles(ctx, c, recipe)
	}

	rb, err := bc.Build(ctx, func(ctx context.Context, platform *ocispecs.Platform, _ int) (client.Reference, *dockerspec.DockerOCIImage, *dockerspec.DockerOCIImage, error) {
		recipe := plan.Recipe(*src)
		base, err := resolveImage(ctx, c, recipe.Base, platform)
		if err != nil {
			return nil, nil, nil, err
		}
		var opts []llb.ConstraintsOpt
		if platform != nil {
			opts = append(opts, llb.Platform(*platform))
		}
		def, err := recipe.State.Marshal(ctx, opts...)
		if err != nil {
			return nil, nil, nil, err
		}
		res, err := c.Solve(ctx, client.SolveRequest{Definition: def.ToPB()})
		if err != nil {
			return nil, nil, nil, err
		}
		ref, err := res.SingleRef()
		if err != nil {
			return nil, nil, nil, err
		}
		return ref, recipe.image(base), base, nil
	})
	if err != nil {
		return nil, err
	}
	return rb.Finalize()
}

// solveFiles returns the recipe's filesystem as the whole result: a ref and no
// metadata. --output type=local (or tar) writes the built site out; with no
// image config there is nothing for an image exporter to make runnable. A site
// is the same files on every architecture, so it is built once.
func solveFiles(ctx context.Context, c client.Client, recipe Recipe) (*client.Result, error) {
	def, err := recipe.State.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	res, err := c.Solve(ctx, client.SolveRequest{Definition: def.ToPB()})
	if err != nil {
		return nil, err
	}
	ref, err := res.SingleRef()
	if err != nil {
		return nil, err
	}
	out := client.NewResult()
	out.SetRef(ref)
	return out, nil
}

// detectContext solves the context to a ref, lists the root, and runs the pure
// Detect over the filenames it finds.
func detectContext(ctx context.Context, c client.Client, src llb.State) (Plan, error) {
	def, err := src.Marshal(ctx)
	if err != nil {
		return Plan{}, err
	}
	res, err := c.Solve(ctx, client.SolveRequest{Definition: def.ToPB()})
	if err != nil {
		return Plan{}, err
	}
	ref, err := res.SingleRef()
	if err != nil {
		return Plan{}, err
	}
	entries, err := ref.ReadDir(ctx, client.ReadDirRequest{Path: "/"})
	if err != nil {
		return Plan{}, err
	}
	files := make(map[string]bool, len(entries))
	for _, e := range entries {
		files[path.Base(e.Path)] = true
	}
	var npm Scripts
	if files["package.json"] {
		// Bounded: the context is untrusted input and a real manifest is
		// kilobytes. Whether the project can be started decides its lane.
		data, err := ref.ReadFile(ctx, client.ReadRequest{
			Filename: "package.json",
			Range:    &client.FileRange{Length: maxPackage},
		})
		if err != nil {
			return Plan{}, err
		}
		if npm, err = scripts(data); err != nil {
			return Plan{}, err
		}
	}
	return Detect(files, npm)
}

// resolveImage fetches a base image's OCI config so its runtime settings
// (notably PATH) are inherited by the produced image.
func resolveImage(ctx context.Context, c client.Client, ref string, platform *ocispecs.Platform) (*dockerspec.DockerOCIImage, error) {
	_, _, data, err := c.ResolveImageConfig(ctx, ref, sourceresolver.Opt{
		Platform: platform,
		ImageOpt: &sourceresolver.ResolveImageOpt{},
	})
	if err != nil {
		return nil, err
	}
	var img dockerspec.DockerOCIImage
	if err := json.Unmarshal(data, &img); err != nil {
		return nil, err
	}
	return &img, nil
}
