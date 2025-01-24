package generate

type StartContext struct {
	BaseImage string
	Command   string
	Paths     []string
	Env       map[string]string
}

func NewStartContext() *StartContext {
	return &StartContext{
		Env: make(map[string]string),
	}
}
