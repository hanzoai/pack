package config

import (
	"github.com/invopop/jsonschema"
	"github.com/railwayapp/railpack/core/plan"
)

type Config struct {
	// The base image to use for the build
	BaseImage string `json:"baseImage,omitempty" jsonschema:"description=The base image to use for the build"`

	// Map of step names to step definitions
	Steps map[string]*plan.Step `json:"steps,omitempty" jsonschema:"description=Map of step names to step definitions"`

	// Start configuration
	Start plan.Start `json:"start,omitempty" jsonschema:"description=Start configuration"`

	// Map of package name to package version
	Packages map[string]string `json:"packages,omitempty" jsonschema:"description=Map of package name to package version"`

	// List of apt packages to install
	AptPackages []string `json:"aptPackages,omitempty" jsonschema:"description=List of apt packages to install"`

	// Map of cache name to cache definitions. The cache key can be referenced in an exec command.
	Caches map[string]*plan.Cache `json:"caches,omitempty" jsonschema:"description=Map of cache name to cache definitions. The cache key can be referenced in an exec command"`
}

func EmptyConfig() *Config {
	return &Config{
		Steps:       make(map[string]*plan.Step),
		Packages:    make(map[string]string),
		AptPackages: make([]string, 0),
		Caches:      make(map[string]*plan.Cache),
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

func (Config) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Properties.Set("$schema", &jsonschema.Schema{
		Type:        "string",
		Description: "The schema for this config",
	})
}

func GetJsonSchema() *jsonschema.Schema {
	r := jsonschema.Reflector{
		DoNotReference: true,
		Anonymous:      true,
	}

	schema := r.Reflect(&Config{})
	return schema
}
