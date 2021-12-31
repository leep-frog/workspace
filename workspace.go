package workspace

import (
	"encoding/json"
	"fmt"

	"github.com/leep-frog/command"
)

const (
	workspaceArg = "WORKSPACE"
)

var (
	nArg  = command.BashCommand(command.IntType, "numWorkspaces", []string{"wmctrl -d | wc | awk '{ print $1 }'"})
	cwArg = command.BashCommand(command.IntType, "currentWorkspace", []string{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`})
)

func CLI() *Workspace {
	return &Workspace{}
}

type Workspace struct {
	Prev    int
	changed bool
}

func (w *Workspace) Load(jsn string) error {
	if jsn == "" {
		w = &Workspace{}
		return nil
	}

	if err := json.Unmarshal([]byte(jsn), w); err != nil {
		return fmt.Errorf("failed to unmarshal json for workspace object: %v", err)
	}
	return nil
}

func (*Workspace) Name() string {
	return "ws"
}

func (w *Workspace) Changed() bool {
	return w.changed
}

func (*Workspace) Setup() []string {
	return nil
}

func (w *Workspace) moveRelative(offset int, output command.Output, data *command.Data) ([]string, error) {
	n := nArg.Get(data).ToInt()
	c := cwArg.Get(data).ToInt()
	if n <= 0 {
		return nil, output.Stderr("couldn't get number of workspaces")
	}
	var newWS int
	for newWS = c + offset; newWS < 0; newWS += n {
	}
	newWS = newWS % n
	return w.moveTo(newWS, output, data)
}

func (w *Workspace) moveTo(n int, output command.Output, data *command.Data) ([]string, error) {
	c := cwArg.Get(data).ToInt()
	// If we're already in the workspace, then just return.
	if n == c {
		return nil, nil
	}
	w.Prev = c
	w.changed = true
	return []string{fmt.Sprintf("wmctrl -s %d", n)}, nil
}

func (w *Workspace) nthWorkspace(output command.Output, data *command.Data) ([]string, error) {
	return w.moveTo(data.Int(workspaceArg), output, data)
}

func (w *Workspace) moveBack(output command.Output, data *command.Data) ([]string, error) {
	return w.moveTo(w.Prev, output, data)
}

func (w *Workspace) moveLeft(output command.Output, data *command.Data) ([]string, error) {
	return w.moveRelative(-1, output, data)
}

func (w *Workspace) moveRight(output command.Output, data *command.Data) ([]string, error) {
	return w.moveRelative(1, output, data)
}

func (w *Workspace) Node() *command.Node {
	return command.BranchNode(
		map[string]*command.Node{
			"left":  command.SerialNodes(nArg, cwArg, command.ExecutableNode(w.moveLeft)),
			"right": command.SerialNodes(nArg, cwArg, command.ExecutableNode(w.moveRight)),
			"back":  command.SerialNodes(cwArg, command.ExecutableNode(w.moveBack)),
		},
		command.SerialNodes(
			cwArg,
			command.IntNode(workspaceArg, "Workspace number", command.IntNonNegative()),
			command.ExecutableNode(w.nthWorkspace),
		),
		true,
	)
}
