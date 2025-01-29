package build_llb

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

	for k, v := range other.EnvVars {
		e.EnvVars[k] = v
	}
}

func (e *BuildEnvironment) AddPath(path string) {
	for _, existingPath := range e.PathList {
		if existingPath == path {
			return
		}
	}

	e.PathList = append(e.PathList, path)
}

func (e *BuildEnvironment) AddEnvVar(key, value string) {
	e.EnvVars[key] = value
}
