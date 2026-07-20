package main

import (
	"strings"
	"testing"
)

func fileset(names ...string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set
}

func TestDetect(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		npm   Scripts
		want  Ecosystem
		err   bool
	}{
		{"go", []string{"go.mod", "go.sum", "main.go"}, Scripts{}, Go, false},
		{"node", []string{"package.json"}, Scripts{}, Node, false},
		{"python pyproject", []string{"pyproject.toml"}, Scripts{}, Python, false},
		{"python requirements", []string{"requirements.txt"}, Scripts{}, Python, false},
		{"static", []string{"index.html", "style.css"}, Scripts{}, Static, false},
		{"go beats node", []string{"go.mod", "package.json"}, Scripts{}, Go, false},
		{"node beats python", []string{"package.json", "requirements.txt"}, Scripts{}, Node, false},
		{"python beats static", []string{"requirements.txt", "index.html"}, Scripts{}, Python, false},
		{"unsupported", []string{"README.md", "LICENSE"}, Scripts{}, "", true},
		{"empty", nil, Scripts{}, "", true},

		// A package.json is a site only when it cannot be started.
		{"generated site", []string{"package.json"}, Scripts{Build: true}, Static, false},
		{"app that builds and starts", []string{"package.json"}, Scripts{Build: true, Start: true}, Node, false},
		{"app that only starts", []string{"package.json"}, Scripts{Start: true}, Node, false},
		{"go beats generated site", []string{"go.mod", "package.json"}, Scripts{Build: true}, Go, false},
		{"generated site beats index.html", []string{"package.json", "index.html"}, Scripts{Build: true}, Static, false},

		// A Dockerfile holds a context in the image lane.
		{"dockerfile keeps a generator an app", []string{"package.json", "Dockerfile"}, Scripts{Build: true}, Node, false},
		{"dockerfile is not a static site", []string{"index.html", "Dockerfile"}, Scripts{}, "", true},
		{"dockerfile alone", []string{"Dockerfile"}, Scripts{}, "", true},
		{"dockerfile does not disturb go", []string{"go.mod", "Dockerfile"}, Scripts{}, Go, false},
		{"dockerfile does not disturb python", []string{"requirements.txt", "Dockerfile"}, Scripts{}, Python, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Detect(fileset(tc.files...), tc.npm)
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got plan %+v", plan)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plan.Ecosystem != tc.want {
				t.Fatalf("ecosystem = %q, want %q", plan.Ecosystem, tc.want)
			}
		})
	}
}

func TestDetectFlags(t *testing.T) {
	if p, _ := Detect(fileset("package.json", "package-lock.json"), Scripts{}); !p.Lockfile {
		t.Error("package-lock.json should set Lockfile (npm ci)")
	}
	if p, _ := Detect(fileset("package.json"), Scripts{}); p.Lockfile {
		t.Error("no lockfile should leave Lockfile false (npm install)")
	}
	if p, _ := Detect(fileset("requirements.txt"), Scripts{}); !p.Requirements {
		t.Error("requirements.txt should set Requirements")
	}
	if p, _ := Detect(fileset("pyproject.toml"), Scripts{}); p.Requirements {
		t.Error("pyproject-only should leave Requirements false")
	}
	p, _ := Detect(fileset("package.json", "package-lock.json"), Scripts{Build: true})
	if !p.Generated || !p.Lockfile {
		t.Errorf("a generated site keeps its lockfile: %+v", p)
	}
	if p, _ := Detect(fileset("index.html"), Scripts{}); p.Generated {
		t.Error("plain HTML is not generated")
	}
}

// The Dockerfile refusal names the escape hatch, since that is the whole point
// of refusing: pack cannot build a Dockerfile, dockerfile.v0 can.
func TestDetectDockerfileNamesTheEscapeHatch(t *testing.T) {
	_, err := Detect(fileset("index.html", "Dockerfile"), Scripts{})
	if err == nil || !strings.Contains(err.Error(), "dockerfile.v0") {
		t.Fatalf("error = %v, want one naming dockerfile.v0", err)
	}
}

func TestScripts(t *testing.T) {
	cases := []struct {
		name string
		json string
		want Scripts
		err  bool
	}{
		{"build only", `{"scripts":{"build":"vite build"}}`, Scripts{Build: true}, false},
		{"build and start", `{"scripts":{"build":"next build","start":"next start"}}`, Scripts{Build: true, Start: true}, false},
		{"start only", `{"scripts":{"start":"node server.js"}}`, Scripts{Start: true}, false},
		{"no scripts", `{"name":"lib","version":"1.0.0"}`, Scripts{}, false},
		{"other scripts", `{"scripts":{"test":"vitest","dev":"vite"}}`, Scripts{}, false},
		{"empty start runs nothing", `{"scripts":{"build":"x","start":""}}`, Scripts{Build: true}, false},
		{"malformed", `{"scripts":`, Scripts{}, true},
		{"scripts not an object", `{"scripts":"nope"}`, Scripts{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := scripts([]byte(tc.json))
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("scripts = %+v, want %+v", got, tc.want)
			}
		})
	}
}
