package main

import "testing"

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
		want  Ecosystem
		err   bool
	}{
		{"go", []string{"go.mod", "go.sum", "main.go"}, Go, false},
		{"node", []string{"package.json"}, Node, false},
		{"python pyproject", []string{"pyproject.toml"}, Python, false},
		{"python requirements", []string{"requirements.txt"}, Python, false},
		{"static", []string{"index.html", "style.css"}, Static, false},
		{"go beats node", []string{"go.mod", "package.json"}, Go, false},
		{"node beats python", []string{"package.json", "requirements.txt"}, Node, false},
		{"python beats static", []string{"requirements.txt", "index.html"}, Python, false},
		{"unsupported", []string{"README.md", "LICENSE"}, "", true},
		{"empty", nil, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Detect(fileset(tc.files...))
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
	if p, _ := Detect(fileset("package.json", "package-lock.json")); !p.Lockfile {
		t.Error("package-lock.json should set Lockfile (npm ci)")
	}
	if p, _ := Detect(fileset("package.json")); p.Lockfile {
		t.Error("no lockfile should leave Lockfile false (npm install)")
	}
	if p, _ := Detect(fileset("requirements.txt")); !p.Requirements {
		t.Error("requirements.txt should set Requirements")
	}
	if p, _ := Detect(fileset("pyproject.toml")); p.Requirements {
		t.Error("pyproject-only should leave Requirements false")
	}
}
