package node

type PackageJson struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Scripts         map[string]string `json:"scripts"`
	PackageManager  *string           `json:"packageManager"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Main            string            `json:"main"`
	Workspaces      []string          `json:"workspaces"`
}

func NewPackageJson() *PackageJson {
	return &PackageJson{
		Scripts:    map[string]string{},
		Engines:    map[string]string{},
		Workspaces: []string{},
	}
}

func (p *PackageJson) HasScript(name string) bool {
	return p.Scripts != nil && p.Scripts[name] != ""
}

func (p *PackageJson) GetScript(name string) string {
	if p.Scripts == nil {
		return ""
	}

	return p.Scripts[name]
}
