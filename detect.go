package main

import (
	"encoding/json"
	"fmt"
)

// Ecosystem is a project kind pack knows how to build.
type Ecosystem string

const (
	Go     Ecosystem = "go"
	Node   Ecosystem = "node"
	Python Ecosystem = "python"
	Static Ecosystem = "static"
)

// Plan is what Detect decides from the files at the context root: the
// ecosystem plus the few facts a Recipe needs.
type Plan struct {
	Ecosystem    Ecosystem
	Lockfile     bool // package-lock.json present -> npm ci
	Requirements bool // python: requirements.txt present -> pip install -r
	Generated    bool // static: the site is built by npm run build
}

// Scripts are the package.json scripts pack reads. A project that can be
// started is an app; one that can only be built is a site.
type Scripts struct {
	Build bool
	Start bool
}

// Detect maps the filenames at the context root, plus the scripts of a
// package.json when there is one, to a build Plan. This order is the single
// source of precedence; the first match wins.
func Detect(files map[string]bool, npm Scripts) (Plan, error) {
	// A Dockerfile is the author declaring an image, so it holds the context in
	// the image lane: the static lane, which emits files and runs no server,
	// never claims it. Every other ecosystem is unaffected by it.
	site := !files["Dockerfile"]
	switch {
	case files["go.mod"]:
		return Plan{Ecosystem: Go}, nil
	case files["package.json"]:
		// A package with a build script and no start script cannot be started,
		// so it is a generated site, not an app: npm start has nothing to run.
		if site && npm.Build && !npm.Start {
			return Plan{Ecosystem: Static, Generated: true, Lockfile: files["package-lock.json"]}, nil
		}
		return Plan{Ecosystem: Node, Lockfile: files["package-lock.json"]}, nil
	case files["pyproject.toml"] || files["requirements.txt"]:
		return Plan{Ecosystem: Python, Requirements: files["requirements.txt"]}, nil
	case site && files["index.html"]:
		return Plan{Ecosystem: Static}, nil
	case files["Dockerfile"]:
		return Plan{}, fmt.Errorf("Dockerfile at context root: build it with frontend=dockerfile.v0")
	}
	return Plan{}, fmt.Errorf("no supported project at context root (looked for go.mod, package.json, pyproject.toml, requirements.txt, index.html)")
}

// scripts reads the script names out of a package.json. A script that is
// present but empty runs nothing, so it does not count.
func scripts(data []byte) (Scripts, error) {
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Scripts{}, fmt.Errorf("package.json: %w", err)
	}
	return Scripts{Build: pkg.Scripts["build"] != "", Start: pkg.Scripts["start"] != ""}, nil
}
