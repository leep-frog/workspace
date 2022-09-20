package workspace

import (
	"fmt"
	"sort"
	"strings"

	"github.com/leep-frog/command"
)

const (
	workspaceArg  = "WORKSPACE"
	monitorArg    = "MONITOR_CODE"
	brightnessArg = "BRIGHTNESS"
)

var (
	nArg = &command.BashCommand[int]{
		ArgName:  "numWorkspaces",
		Contents: []string{"wmctrl -d | wc | awk '{ print $1 }'"},
	}
	cwArg = &command.BashCommand[int]{
		ArgName:  "currentWorkspace",
		Contents: []string{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
	}

	listMcs = &command.BashCommand[[]string]{
		ArgName:  "mcs",
		Contents: []string{`xrandr --query | grep "\bconnected" | awk '{print $1}' | grep -v ^\s*$`},
	}
)

func CLI() *Workspace {
	return &Workspace{}
}

type Workspace struct {
	Prev       int
	Brightness map[int]int
	changed    bool
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
	n := nArg.Get(data)
	c := cwArg.Get(data)
	if n <= 0 {
		return nil, output.Stderrln("couldn't get number of workspaces")
	}
	var newWS int
	for newWS = c + offset; newWS < 0; newWS += n {
	}
	newWS = newWS % n
	return w.moveTo(newWS, output, data)
}

func (w *Workspace) moveTo(n int, output command.Output, data *command.Data) ([]string, error) {
	c := cwArg.Get(data)
	// If we're already in the workspace, then just return.
	if n == c {
		return nil, nil
	}
	w.Prev = c
	w.changed = true
	r := []string{
		fmt.Sprintf("wmctrl -s %d", n),
	}
	b, ok := w.Brightness[n]
	if !ok {
		b = 100
	}
	mcs, err := listMcs.Run(output, data)
	if err != nil {
		output.Annotate(err, "Failed to get monitor codes")
	} else {
		r = append(r, setBrightness(mcs, b)...)
	}
	return r, nil
}

func setBrightness(mcs []string, brightness int) []string {
	var r []string
	for _, mc := range mcs {
		r = append(r, fmt.Sprintf("xrandr --output %s --brightness %0.2f", strings.TrimSpace(mc), float64(brightness)/100.0))
	}
	return r
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

func (w *Workspace) offsetBrightness(offset int) func(o command.Output, d *command.Data) ([]string, error) {
	return func(o command.Output, d *command.Data) ([]string, error) {
		cw := cwArg.Get(d)
		b := 100
		if eb, ok := w.Brightness[cw]; ok {
			b = eb
		}
		b += offset
		if w.Brightness == nil {
			w.Brightness = map[int]int{}
		}
		w.Brightness[cw] = b
		w.changed = true
		return setBrightness(listMcs.Get(d), b), nil
	}
}

func (w *Workspace) Node() *command.Node {
	wn := command.Arg[int](workspaceArg, "Workspace number", command.NonNegative[int]())
	return command.AsNode(&command.BranchNode{
		Branches: map[string]*command.Node{
			"left":  command.SerialNodes(command.Description("Move one workspace left"), nArg, cwArg, command.ExecutableNode(w.moveLeft)),
			"right": command.SerialNodes(command.Description("Move one workspace right"), nArg, cwArg, command.ExecutableNode(w.moveRight)),
			"back":  command.SerialNodes(command.Description("Move to the previous"), cwArg, command.ExecutableNode(w.moveBack)),
			"monitors": command.AsNode(&command.BranchNode{
				Branches: map[string]*command.Node{
					"list": command.SerialNodes(
						command.Description("List monitor codes"),
						listMcs,
						&command.ExecutorProcessor{F: func(o command.Output, d *command.Data) error {
							codes := d.StringList("mcs")
							sort.Strings(codes)
							for _, c := range codes {
								o.Stdoutln(c)
							}
							return nil
						}},
					),
				},
			}),
			"brightness": command.AsNode(&command.BranchNode{
				Branches: map[string]*command.Node{
					"up": command.SerialNodes(
						cwArg,
						listMcs,
						command.ExecutableNode(w.offsetBrightness(10)),
					),
					"down": command.SerialNodes(
						cwArg,
						listMcs,
						command.ExecutableNode(w.offsetBrightness(-10)),
					),
					"set": command.SerialNodes(
						command.Description("Set the brightness for a workspace"),
						wn,
						command.Arg[int](brightnessArg, "Monitor brightness", command.GTE[int](5), command.LTE[int](250)),
						&command.ExecutorProcessor{F: func(o command.Output, d *command.Data) error {
							if w.Brightness == nil {
								w.Brightness = map[int]int{}
							}
							w.Brightness[d.Int(workspaceArg)] = d.Int(brightnessArg)
							w.changed = true
							return nil
						}},
					),
					"list": command.SerialNodes(
						command.Description("List brightnesses for each workspace"),
						&command.ExecutorProcessor{F: func(o command.Output, d *command.Data) error {
							var keys []int
							for k := range w.Brightness {
								keys = append(keys, k)
							}
							sort.Ints(keys)
							for _, k := range keys {
								o.Stdoutf("%2d: %d\n", k, w.Brightness[k])
							}
							return nil
						}},
					),
				},
			}),
		},
		Default: command.SerialNodes(
			command.Description("Move to a specific workspace"),
			wn,
			cwArg,
			command.ExecutableNode(w.nthWorkspace),
		),
	})
}
