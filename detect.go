package main

import "fmt"

// Ecosystem is a project kind pack knows how to build.
type Ecosystem string

const (
	Go     Ecosystem = "go"
	Node   Ecosystem = "node"
	Python Ecosystem = "python"
	Static Ecosystem = "static"
)

// Plan is what Detect decides from the files at the context root: the
// ecosystem plus the few filename facts a Recipe needs.
type Plan struct {
	Ecosystem    Ecosystem
	Lockfile     bool // node: package-lock.json present -> npm ci
	Requirements bool // python: requirements.txt present -> pip install -r
}

// Detect maps the set of filenames at the context root to a build Plan. This
// order is the single source of precedence; the first match wins.
func Detect(files map[string]bool) (Plan, error) {
	switch {
	case files["go.mod"]:
		return Plan{Ecosystem: Go}, nil
	case files["package.json"]:
		return Plan{Ecosystem: Node, Lockfile: files["package-lock.json"]}, nil
	case files["pyproject.toml"] || files["requirements.txt"]:
		return Plan{Ecosystem: Python, Requirements: files["requirements.txt"]}, nil
	case files["index.html"]:
		return Plan{Ecosystem: Static}, nil
	}
	return Plan{}, fmt.Errorf("no supported project at context root (looked for go.mod, package.json, pyproject.toml, requirements.txt, index.html)")
}
