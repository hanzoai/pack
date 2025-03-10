package build_llb

import (
	"maps"
	"slices"
)

type BuildEnvironment struct {
	PathList []string
	EnvVars  map[string]string
}

func NewGraphEnvironment() BuildEnvironment {
	return BuildEnvironment{
		PathList: make([]string, 0),
		EnvVars:  make(map[string]string),
	}
}

// Merges the other environment into the current environment
func (e *BuildEnvironment) Merge(other BuildEnvironment) {
	e.PathList = append(e.PathList, other.PathList...)
	maps.Copy(e.EnvVars, other.EnvVars)
}

func (e *BuildEnvironment) PushPath(path string) {
	if slices.Contains(e.PathList, path) {
		return
	}
	e.PathList = append([]string{path}, e.PathList...)
}

func (e *BuildEnvironment) AddEnvVar(key, value string) {
	e.EnvVars[key] = value
}
