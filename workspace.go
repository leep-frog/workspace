package workspace

import (
	"fmt"

	"github.com/leep-frog/command"
)

const (
	workspaceArg = "workspace"
)

var (
	runInt = command.RunInt
)

func CLI() *Workspace {
	return &Workspace{}
}

type Workspace struct{}

func (*Workspace) Load(jsn string) error {
	return nil
}

func (*Workspace) Name() string {
	return "ws"
}

func (*Workspace) Changed() bool {
	return false
}

func (*Workspace) Setup() []string {
	return nil
}

func numWorkspaces() (int, error, int) {
	return runInt([]string{"wmctrl -d | wc | awk '{ print $1 }'"})
}

func currentWorkspace() (int, error, int) {
	return runInt([]string{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`})
}

func moveWorkspace(offset int, output command.Output, eData *command.ExecuteData) error {
	n, err, _ := numWorkspaces()
	if err != nil {
		return output.Stderr("failed to get number of workspaces: %v", err)
	}
	c, err, _ := currentWorkspace()
	if err != nil {
		return output.Stderr("failed to get current workspace: %v", err)
	}
	var newWS int
	for newWS = c + offset; newWS < 0; newWS += n {
	}
	newWS = newWS % n
	eData.Executable = append(eData.Executable, fmt.Sprintf("wmctrl -s %d", newWS))
	return nil
}

func (w *Workspace) moveLeft(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return moveWorkspace(-1, output, eData)
}

func (w *Workspace) moveRight(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return moveWorkspace(1, output, eData)
}

// TODO: in command package: "func SimpleExecutable(f func(data, eData) error)"
func (w *Workspace) nthWorkspace(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	eData.Executable = append(eData.Executable, fmt.Sprintf("wmctrl -s %d", data.Int(workspaceArg)))
	return nil
}

func (w *Workspace) Node() *command.Node {
	return command.BranchNode(
		map[string]*command.Node{
			"left":  command.SerialNodes(command.SimpleProcessor(w.moveLeft, nil)),
			"right": command.SerialNodes(command.SimpleProcessor(w.moveRight, nil)),
		},
		command.SerialNodes(
			// TODO: change ArgOpt to set of options.
			// TODO: remove "Node" from IntNode, StringNode, FloatNode, BoolNode.
			command.IntNode(workspaceArg, command.IntNonNegative()),
			command.SimpleProcessor(w.nthWorkspace, nil),
		),
		true,
	)
}
