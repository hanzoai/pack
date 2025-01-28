package config

import (
	"github.com/invopop/jsonschema"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/utils"
)

type Config struct {
	// The base image to use for the build
	BaseImage string `json:"baseImage,omitempty" jsonschema:"description=The base image to use for the build"`

	// List of providers to use
	Providers *[]string `json:"providers,omitempty" jsonschema:"description=List of providers to use"`

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
		Steps:    make(map[string]*plan.Step),
		Packages: make(map[string]string),
		Caches:   make(map[string]*plan.Cache),
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

// Merge combines multiple configs by merging their values with later configs taking precedence
func Merge(configs ...*Config) *Config {
	if len(configs) == 0 {
		return EmptyConfig()
	}

	result := EmptyConfig()

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Strings (use last non-empty value)
		if config.BaseImage != "" {
			result.BaseImage = config.BaseImage
		}

		if config.Start.Command != "" {
			result.Start = config.Start
		}

		// Maps (overwrite existing values)
		for k, v := range config.Caches {
			result.Caches[k] = v
		}
		for k, v := range config.Packages {
			result.Packages[k] = v
		}
		for k, v := range config.Steps {
			result.Steps[k] = v
		}

		// Arrays (extend)
		result.AptPackages = append(result.AptPackages, config.AptPackages...)

		// Merge providers
		result.Providers = utils.MergeStringSlicePointers(result.Providers, config.Providers)
	}

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
