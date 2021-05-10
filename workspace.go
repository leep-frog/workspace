package workspace

import (
	"strconv"

	"github.com/leep-frog/command"
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

func numWorkspaces() (int, error) {
	return runInt([]string{"wmctrl -d | wc | awk '{ print $1 }'"})
}

func currentWorkspace() (int, error) {
	return runInt([]string{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`})
}

func moveWorkspace(offset int, output command.Output, eData *command.ExecuteData) error {
	n, err := numWorkspaces()
	if err != nil {
		return output.Stderr("failed to get number of workspaces: %v", err)
	}
	c, err := currentWorkspace()
	if err != nil {
		return output.Stderr("failed to get current workspace: %v", err)
	}
	var newWS int
	for newWS = c + offset; newWS < 0; newWS += n {
	}
	newWS = newWS % n
	eData.Executable = append(eData.Executable, []string{"wmctrl", "-s", strconv.Itoa(newWS)})
	return nil
}

func (w *Workspace) moveLeft(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return moveWorkspace(-1, output, eData)
}

func (w *Workspace) moveRight(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return moveWorkspace(1, output, eData)
}

func (w *Workspace) Node() *command.Node {
	return command.BranchNode(
		map[string]*command.Node{
			"left":  command.SerialNodes(command.SimpleProcessor(w.moveLeft, nil)),
			"right": command.SerialNodes(command.SimpleProcessor(w.moveRight, nil)),
		},
		nil,
		true,
	)
}
