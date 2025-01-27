package config

import "github.com/railwayapp/railpack-go/core/plan"

type Config struct {
	BaseImage   string                `json:"baseImage,omitempty"`
	Caches      map[string]plan.Cache `json:"caches,omitempty"`
	Packages    map[string]string     `json:"packages,omitempty"`
	AptPackages []string              `json:"aptPackages,omitempty"`
	Steps       []plan.Step           `json:"steps,omitempty"`
	Start       plan.Start            `json:"start,omitempty"`
}

func EmptyConfig() *Config {
	return &Config{
		Caches:      make(map[string]plan.Cache),
		Packages:    make(map[string]string),
		AptPackages: make([]string, 0),
		Steps:       make([]plan.Step, 0),
	}
}

func (c *Config) SetBuildCommand(cmd string) {
	buildStep := c.GetOrCreateStep("build")
	buildStep.Commands = []plan.Command{plan.NewExecCommand(cmd)}
}

func (c *Config) SetInstallCommand(cmd string) {
	installStep := c.GetOrCreateStep("install")
	installStep.Commands = []plan.Command{plan.NewExecCommand(cmd)}
}

func (c *Config) GetOrCreateStep(name string) *plan.Step {
	for _, step := range c.Steps {
		if step.Name == name {
			return &step
		}
	}

	step := plan.NewStep(name)
	c.Steps = append(c.Steps, *step)
	return step
}

// Merge combines two configs where:
// - For strings (BaseImage), the last value wins
// - For maps (Caches, Packages), entries are merged with last value winning
// - For arrays (AptPackages, Steps), arrays are extended
func (c *Config) Merge(other *Config) *Config {
	result := &Config{
		Caches:      make(map[string]plan.Cache),
		Packages:    make(map[string]string),
		AptPackages: make([]string, 0),
		Steps:       make([]plan.Step, 0),
	}

	// Copy maps from first config
	for k, v := range c.Caches {
		result.Caches[k] = v
	}
	for k, v := range c.Packages {
		result.Packages[k] = v
	}

	// Copy arrays from first config
	result.AptPackages = append(result.AptPackages, c.AptPackages...)
	result.Steps = append(result.Steps, c.Steps...)

	// Merge in second config
	if other.BaseImage != "" {
		result.BaseImage = other.BaseImage
	} else {
		result.BaseImage = c.BaseImage
	}

	// Handle Start field (last non-empty value wins)
	if other.Start.Command != "" {
		result.Start = other.Start
	} else {
		result.Start = c.Start
	}

	// Merge maps from second config (overwriting existing values)
	for k, v := range other.Caches {
		result.Caches[k] = v
	}
	for k, v := range other.Packages {
		result.Packages[k] = v
	}

	// Extend arrays from second config
	result.AptPackages = append(result.AptPackages, other.AptPackages...)
	result.Steps = append(result.Steps, other.Steps...)

	return result
}
