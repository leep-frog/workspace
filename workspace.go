package workspace

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/leep-frog/command"
)

const (
	workspaceArg  = "WORKSPACE"
	monitorArg    = "MONITOR_CODE"
	brightnessArg = "BRIGHTNESS"
)

var (
	nArg  = command.BashCommand(command.IntType, "numWorkspaces", []string{"wmctrl -d | wc | awk '{ print $1 }'"})
	cwArg = command.BashCommand(command.IntType, "currentWorkspace", []string{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`})
)

func CLI() *Workspace {
	return &Workspace{}
}

type Workspace struct {
	Prev       int
	Brightness map[int]int
	changed    bool
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

func (w *Workspace) setBrightness(output command.Output, data *command.Data) ([]string, error) {
	if w.Brightness == nil {
		w.Brightness = map[int]int{}
	}
	//w.Brightness[data.Int(workspaceArg)] =
	return w.moveRelative(1, output, data)
}

func (w *Workspace) Node() *command.Node {
	wn := command.IntNode(workspaceArg, "Workspace number", command.IntNonNegative())
	return command.BranchNode(
		map[string]*command.Node{
			"left":  command.SerialNodes(command.Description("Move one workspace left"), nArg, cwArg, command.ExecutableNode(w.moveLeft)),
			"right": command.SerialNodes(command.Description("Move one workspace right"), nArg, cwArg, command.ExecutableNode(w.moveRight)),
			"back":  command.SerialNodes(command.Description("Move to the previous"), cwArg, command.ExecutableNode(w.moveBack)),
			"monitors": command.BranchNode(map[string]*command.Node{
				"list": command.SerialNodes(
					command.Description("List monitor codes"),
					command.BashCommand(command.StringListType, "mcs", []string{`xrandr --query | grep "\bconnected" | awk '{print $1}'`}),
					command.ExecutorNode(func(o command.Output, d *command.Data) {
						codes := d.StringList("mcs")
						sort.Strings(codes)
						for _, c := range codes {
							o.Stdoutln(c)
						}
					}),
				),
			}, nil),
			"brightness": command.BranchNode(map[string]*command.Node{
				"set": command.SerialNodes(
					command.Description("Set the brightness for a workspace"),
					wn,
					command.IntNode(brightnessArg, "Monitor brightness", command.IntGTE(5), command.IntLTE(250)),
					command.ExecutorNode(func(o command.Output, d *command.Data) {
						if w.Brightness == nil {
							w.Brightness = map[int]int{}
						}
						w.Brightness[d.Int(workspaceArg)] = d.Int(brightnessArg)
						w.changed = true
					}),
				),
				"list": command.SerialNodes(
					command.Description("List brightnesses for each workspace"),
					command.ExecutorNode(func(o command.Output, d *command.Data) {
						var keys []int
						for k := range w.Brightness {
							keys = append(keys, k)
						}
						sort.Ints(keys)
						for _, k := range keys {
							o.Stdoutf("%2d: %d", k, w.Brightness[k])
						}
					}),
				),
			}, nil),
		},
		command.SerialNodes(
			command.Description("Move to a specific workspace"),
			wn,
			cwArg,
			command.ExecutableNode(w.nthWorkspace),
		),
	)
}
