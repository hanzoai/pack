package plan

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

type Step struct {
	// The name of the step
	Name string `json:"name,omitempty" jsonschema:"description=The name of the step"`

	// The steps that this step depends on. The step will only run after all the steps in DependsOn have run
	DependsOn []string `json:"dependsOn,omitempty" jsonschema:"description=The steps that this step depends on. The step will only run after all the steps in DependsOn have run"`

	// The commands to run in this step
	Commands []Command `json:"commands,omitempty" jsonschema:"description=The commands to run in this step"`

	// Paths that this step outputs. Only these paths will be available to the next step
	Outputs []string `json:"outputs,omitempty" jsonschema:"description=Paths that this step outputs. Only these paths will be available to the next step"`

	// The assets available to this step. The key is the name of the asset that is referenced in a file command
	Assets map[string]string `json:"assets,omitempty" jsonschema:"description=The assets available to this step. The key is the name of the asset that is referenced in a file command"`

	// The base image that will be used for this step
	// If empty (default), the base image will be the one from the previous step
	// Only set this if you don't want to reuse any part of the file system from the previous step
	StartingImage string `json:"startingImage,omitempty" jsonschema:"description=The base image that will be used for this step. If empty (default), the base image will be the one from the previous step. Only set this if you don't want to reuse any part of the file system from the previous step"`
}

func NewStep(name string) *Step {
	return &Step{
		Name:      name,
		DependsOn: make([]string, 0),
		Commands:  make([]Command, 0),
	}
}

func (s *Step) DependOn(name string) {
	s.DependsOn = append(s.DependsOn, name)
}

func (s *Step) AddCommands(commands []Command) {
	s.Commands = append(s.Commands, commands...)
}

func MergeSteps(steps ...*Step) *Step {
	if len(steps) == 0 {
		return nil
	}

	result := &Step{
		Name:      steps[0].Name,
		DependsOn: make([]string, 0),
		Commands:  make([]Command, 0),
	}

	for _, step := range steps {
		result.DependsOn = append(result.DependsOn, step.DependsOn...)
		result.Commands = append(result.Commands, step.Commands...)
	}

	return result
}

func (s *Step) UnmarshalJSON(data []byte) error {
	type Alias Step
	aux := &struct {
		Commands []json.RawMessage `json:"commands"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Commands = make([]Command, len(aux.Commands))
	for i, rawCmd := range aux.Commands {
		cmd, err := UnmarshalCommand(rawCmd)
		if err != nil {
			return err
		}
		s.Commands[i] = cmd
	}

	return nil
}

func (Step) JSONSchemaExtend(schema *jsonschema.Schema) {
	// Remove name from the schema
	var required []string
	for _, prop := range schema.Required {
		if prop != "name" {
			required = append(required, "name")
		}
	}
	schema.Required = required
	schema.Properties.Delete("name")

	// Add proper schemas for the commands
	var commandsDescription string
	if currCommandsSchema, ok := schema.Properties.Get("commands"); ok {
		commandsDescription = currCommandsSchema.Description
	}

	commandSchema := &jsonschema.Schema{
		Type:        "array",
		Description: commandsDescription,
		Items:       CommandsSchema(),
	}

	schema.Properties.Set("commands", commandSchema)
}

func CommandsSchema() *jsonschema.Schema {
	execSchema := generateSchemaWithComments(ExecCommand{})
	pathSchema := generateSchemaWithComments(PathCommand{})
	variableSchema := generateSchemaWithComments(VariableCommand{})
	copySchema := generateSchemaWithComments(CopyCommand{})
	fileSchema := generateSchemaWithComments(FileCommand{})

	availableCommands := []*jsonschema.Schema{execSchema, pathSchema, variableSchema, copySchema, fileSchema}

	// Add string schema type as an additional valid command type
	stringSchema := &jsonschema.Schema{
		Type:        "string",
		Description: "Strings will be parsed and interpreted as a command to run",
	}
	availableCommands = append([]*jsonschema.Schema{stringSchema}, availableCommands...)

	return &jsonschema.Schema{
		OneOf: availableCommands,
	}
}

func generateSchemaWithComments(v any) *jsonschema.Schema {
	r := jsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	return r.Reflect(v)
}
