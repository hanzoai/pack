package buildkit

type GraphEnvironment struct {
	PathList []string
	EnvVars  map[string]string
}

func NewGraphEnvironment() GraphEnvironment {
	return GraphEnvironment{
		PathList: make([]string, 0),
		EnvVars:  make(map[string]string),
	}
}

// Merges the other environment into the current environment
func (e *GraphEnvironment) Merge(other GraphEnvironment) {
	e.PathList = append(e.PathList, other.PathList...)

	for k, v := range other.EnvVars {
		e.EnvVars[k] = v
	}
}

func (e *GraphEnvironment) AddPath(path string) {
	for _, existingPath := range e.PathList {
		if existingPath == path {
			return
		}
	}

	e.PathList = append(e.PathList, path)
}

func (e *GraphEnvironment) AddEnvVar(key, value string) {
	e.EnvVars[key] = value
}
