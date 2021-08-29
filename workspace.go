package workspace

import (
	"encoding/json"
	"fmt"

	"github.com/leep-frog/command"
)

const (
	workspaceArg = "workspace"
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

func (w *Workspace) moveRelative(offset int, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	n := nArg.Get(data).Int()
	c := cwArg.Get(data).Int()
	if n <= 0 {
		return output.Stderr("couldn't get number of workspaces")
	}
	var newWS int
	for newWS = c + offset; newWS < 0; newWS += n {
	}
	newWS = newWS % n
	w.moveTo(newWS, output, data, eData)
	return nil
}

func (w *Workspace) moveTo(n int, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	c := cwArg.Get(data).Int()
	// If we're already in the workspace, then just return.
	if n == c {
		return nil
	}
	eData.Executable = append(eData.Executable, fmt.Sprintf("wmctrl -s %d", n))
	w.Prev = c
	w.changed = true
	return nil
}

// TODO: in command package: "func SimpleExecutable(f func(data, eData) error)"
func (w *Workspace) nthWorkspace(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return w.moveTo(data.Int(workspaceArg), output, data, eData)
}

func (w *Workspace) moveBack(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return w.moveTo(w.Prev, output, data, eData)
}

func (w *Workspace) moveLeft(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return w.moveRelative(-1, output, data, eData)
}

func (w *Workspace) moveRight(input *command.Input, output command.Output, data *command.Data, eData *command.ExecuteData) error {
	return w.moveRelative(1, output, data, eData)
}

func (w *Workspace) Node() *command.Node {
	fmt.Println(nArg)
	return command.BranchNode(
		map[string]*command.Node{
			"left":  command.SerialNodes(nArg, cwArg, command.SimpleProcessor(w.moveLeft, nil)),
			"right": command.SerialNodes(nArg, cwArg, command.SimpleProcessor(w.moveRight, nil)),
			"back":  command.SerialNodes(cwArg, command.SimpleProcessor(w.moveBack, nil)),
		},
		command.SerialNodes(
			cwArg,
			command.IntNode(workspaceArg, command.IntNonNegative()),
			command.SimpleProcessor(w.nthWorkspace, nil),
		),
		true,
	)
}
