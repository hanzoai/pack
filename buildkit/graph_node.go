package buildkit

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/railwayapp/railpack/core/plan"
)

type Node struct {
	Step       *plan.Step
	State      *llb.State
	Parents    []*Node
	Children   []*Node
	Processed  bool
	InProgress bool

	InputPathList []string
	InputEnvVars  map[string]string

	OutputPathList []string
	OutputEnvVars  map[string]string
}

func (node *Node) getPathList() []string {
	pathList := make([]string, 0)
	pathList = append(pathList, node.InputPathList...)
	pathList = append(pathList, node.OutputPathList...)
	return pathList
}

func (node *Node) appendPath(path string) {
	// Check if path already exists in input or output paths
	for _, existingPath := range node.OutputPathList {
		if existingPath == path {
			return
		}
	}

	node.OutputPathList = append(node.OutputPathList, path)
}
