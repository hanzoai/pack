package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Command represents a command that can be executed
type Command interface {
	CommandType() string
}

type ExecOptions struct {
	Caches     []string
	CustomName string
}

// ExecCommand represents a command to be executed
type ExecCommand struct {
	Cmd        string   `json:"cmd"`
	Caches     []string `json:"caches,omitempty"`
	CustomName string   `json:"customName,omitempty"`
}

// PathCommand represents a global path addition
type PathCommand struct {
	Path string `json:"path"`
}

// VariableCommand represents a shell variable setting
type VariableCommand struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CopyCommand represents a file copy operation
type CopyCommand struct {
	Image string `json:"image,omitempty"`
	Src   string `json:"src"`
	Dst   string `json:"dst"`
}

type FileOptions struct {
	Mode       os.FileMode `json:"mode,omitempty"`
	CustomName string      `json:"customName,omitempty"`
}

// FileCommand represents a file creation operation
type FileCommand struct {
	Path       string      `json:"path"`
	Name       string      `json:"name"`
	Mode       os.FileMode `json:"mode,omitempty"`
	CustomName string      `json:"customName,omitempty"`
}

func (e ExecCommand) CommandType() string     { return "exec" }
func (g PathCommand) CommandType() string     { return "globalPath" }
func (v VariableCommand) CommandType() string { return "variable" }
func (c CopyCommand) CommandType() string     { return "copy" }
func (f FileCommand) CommandType() string     { return "file" }

func NewExecCommand(cmd string, options ...ExecOptions) Command {
	exec := ExecCommand{Cmd: cmd}
	if len(options) > 0 {
		exec.CustomName = options[0].CustomName
		exec.Caches = options[0].Caches
	}
	return exec
}

func NewPathCommand(path string, customName ...string) Command {
	pathCmd := PathCommand{Path: path}
	return pathCmd
}

func NewVariableCommand(name, value string, customName ...string) Command {
	variableCmd := VariableCommand{Name: name, Value: value}
	return variableCmd
}

func NewCopyCommand(src string, dst ...string) Command {
	dstPath := src
	if len(dst) > 0 {
		dstPath = dst[0]
	}

	copyCmd := CopyCommand{Src: src, Dst: dstPath}
	return copyCmd
}

func NewFileCommand(path, name string, options ...FileOptions) Command {
	fileCmd := FileCommand{Path: path, Name: name}
	if len(options) > 0 {
		fileCmd.CustomName = options[0].CustomName
		fileCmd.Mode = options[0].Mode
	}
	return fileCmd
}

func UnmarshalCommand(data []byte) (Command, error) {
	// First try to unmarshal as JSON object
	if cmd, err := UnmarshalJsonCommand(data); err == nil {
		return cmd, nil
	}

	// If that fails, parse the string into a command
	return UnmarshalStringCommand(data)
}

func UnmarshalJsonCommand(data []byte) (Command, error) {
	// Try to unmarshal as JSON object
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, err
	}

	// Determine command type based on fields present
	if _, ok := rawMap["cmd"]; ok {
		var cmd ExecCommand
		if err := json.Unmarshal(data, &cmd); err != nil {
			return nil, err
		}
		return cmd, nil
	}
	if _, ok := rawMap["path"]; ok {
		if _, ok := rawMap["name"]; ok {
			var file FileCommand
			if err := json.Unmarshal(data, &file); err != nil {
				return nil, err
			}
			return file, nil
		}
		var path PathCommand
		if err := json.Unmarshal(data, &path); err != nil {
			return nil, err
		}
		return path, nil
	}
	if _, ok := rawMap["name"]; ok && rawMap["value"] != nil {
		var env VariableCommand
		if err := json.Unmarshal(data, &env); err != nil {
			return nil, err
		}
		return env, nil
	}
	if _, ok := rawMap["src"]; ok {
		var copy CopyCommand
		if err := json.Unmarshal(data, &copy); err != nil {
			return nil, err
		}
		return copy, nil
	}

	return nil, fmt.Errorf("unknown command type: %v", rawMap)
}

func UnmarshalStringCommand(data []byte) (Command, error) {
	str := string(data)

	// If no prefix, treat as exec command
	if !strings.Contains(str, ":") {
		return NewExecCommand(str), nil
	}

	parts := strings.SplitN(str, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid command format: %s", str)
	}

	prefix := parts[0]
	payload := parts[1]

	// Split prefix into command type and custom name
	prefixParts := strings.SplitN(prefix, "#", 2)
	cmdType := prefixParts[0]
	customName := ""
	if len(prefixParts) > 1 {
		customName = prefixParts[1]
	}

	switch cmdType {
	case "RUN":
		return NewExecCommand(payload, ExecOptions{CustomName: customName}), nil
	case "PATH":
		return NewPathCommand(payload), nil
	case "ENV":
		envParts := strings.SplitN(payload, "=", 2)
		if len(envParts) != 2 {
			return nil, fmt.Errorf("invalid ENV format: %s", payload)
		}
		return NewVariableCommand(envParts[0], envParts[1]), nil
	case "COPY":
		copyParts := strings.Fields(payload)
		if len(copyParts) != 2 {
			return nil, fmt.Errorf("invalid COPY format: %s", payload)
		}
		return NewCopyCommand(copyParts[0], copyParts[1]), nil
	case "FILE":
		fileParts := strings.Fields(payload)
		if len(fileParts) != 2 {
			return nil, fmt.Errorf("invalid FILE format: %s", payload)
		}
		return NewFileCommand(fileParts[0], fileParts[1], FileOptions{CustomName: customName}), nil
	}

	// fallback to exec command type
	return NewExecCommand(str, ExecOptions{CustomName: customName}), nil
}
