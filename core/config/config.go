package config

import (
	"github.com/railwayapp/railpack-go/core/plan"
)

type Config struct {
	BaseImage   string                 `json:"baseImage,omitempty"`
	Caches      map[string]*plan.Cache `json:"caches,omitempty"`
	Packages    map[string]string      `json:"packages,omitempty"`
	AptPackages []string               `json:"aptPackages,omitempty"`
	Steps       map[string]*plan.Step  `json:"steps,omitempty"`
	Start       plan.Start             `json:"start,omitempty"`
}

func EmptyConfig() *Config {
	return &Config{
		Caches:      make(map[string]*plan.Cache),
		Packages:    make(map[string]string),
		AptPackages: make([]string, 0),
		Steps:       make(map[string]*plan.Step),
	}
}

func (c *Config) GetOrCreateStep(name string) *plan.Step {
	step := plan.NewStep(name)
	if existingStep, exists := c.Steps[name]; exists {
		step = existingStep
	}
	c.Steps[name] = step

	return step
}

// Merge combines two configs where:
// - For strings (BaseImage), the last value wins
// - For maps (Caches, Packages, Steps), entries are merged with last value winning
// - For arrays (AptPackages), arrays are extended
func (c *Config) Merge(other *Config) *Config {
	result := EmptyConfig()

	// Copy maps from first config
	for k, v := range c.Caches {
		result.Caches[k] = v
	}
	for k, v := range c.Packages {
		result.Packages[k] = v
	}
	for k, v := range c.Steps {
		result.Steps[k] = v
	}

	// Copy arrays from first config
	result.AptPackages = append(result.AptPackages, c.AptPackages...)

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
	for k, v := range other.Steps {
		result.Steps[k] = v
	}

	// Extend arrays from second config
	result.AptPackages = append(result.AptPackages, other.AptPackages...)

	return result
}
