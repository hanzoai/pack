package generate

type StartContext struct {
	BaseImage string
	Command   string
	outputs   []string
	paths     []string
	variables map[string]string
}

func NewStartContext() *StartContext {
	return &StartContext{
		variables: make(map[string]string),
	}
}

func (s *StartContext) AddEnvVars(envVars map[string]string) {
	for k, v := range envVars {
		s.variables[k] = v
	}
}

func (s *StartContext) AddPaths(paths []string) {
	s.paths = append(s.paths, paths...)
}

func (s *StartContext) AddOutputs(outputs []string) {
	s.outputs = append(s.outputs, outputs...)
}
