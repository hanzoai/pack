package node

type PackageJson struct {
	Scripts         map[string]string `json:"scripts"`
	PackageManager  *string           `json:"packageManager"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Main            *string           `json:"main"`
}

func NewPackageJson() *PackageJson {
	return &PackageJson{
		Scripts: map[string]string{},
		Engines: map[string]string{},
	}
}

func (p *PackageJson) HasScript(name string) bool {
	return p.Scripts != nil && p.Scripts[name] != ""
}
