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

	InputEnv  GraphEnvironment
	OutputEnv GraphEnvironment
}

func (node *Node) getPathList() []string {
	pathList := make([]string, 0)
	pathList = append(pathList, node.InputEnv.PathList...)
	pathList = append(pathList, node.OutputEnv.PathList...)
	return pathList
}

func (node *Node) appendPath(path string) {
	node.OutputEnv.AddPath(path)
}
