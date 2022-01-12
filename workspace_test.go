package workspace

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/leep-frog/command"
)

type runIntResponse struct {
	i   int
	err error
}

func nRun(n int) *command.FakeRun {
	return &command.FakeRun{
		Stdout: []string{fmt.Sprintf("%d", n)},
	}
}

func errRun(s string) *command.FakeRun {
	return &command.FakeRun{
		Err: fmt.Errorf("%s", s),
	}
}

func mcRun(mcs ...string) *command.FakeRun {
	return &command.FakeRun{
		Stdout: mcs,
	}
}

func TestWorkspace(t *testing.T) {
	numW := []string{"set -e", "set -o pipefail", fmt.Sprintf("wmctrl -d | wc | awk '{ print $1 }'")}
	cw := []string{"set -e", "set -o pipefail", fmt.Sprintf(`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`)}
	lmCmd := []string{
		"set -e",
		"set -o pipefail",
		`xrandr --query | grep "\bconnected" | awk '{print $1}'`,
	}

	for _, test := range []struct {
		name string
		w    *Workspace
		etc  *command.ExecuteTestCase
		want *Workspace
	}{
		{
			name: "requires argument",
			etc: &command.ExecuteTestCase{
				WantErr:    fmt.Errorf(`Argument "WORKSPACE" requires at least 1 argument, got 0`),
				WantStderr: []string{`Argument "WORKSPACE" requires at least 1 argument, got 0`},
			},
		},
		{
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:       []string{"up"},
				WantErr:    fmt.Errorf(`strconv.Atoi: parsing "up": invalid syntax`),
				WantStderr: []string{`strconv.Atoi: parsing "up": invalid syntax`},
			},
		},
		{
			name: "fails if runInt fails when getting the number of workspaces",
			etc: &command.ExecuteTestCase{
				RunResponses:    []*command.FakeRun{errRun("unlimited workspaces")},
				Args:            []string{"left"},
				WantErr:         fmt.Errorf("failed to execute bash command: unlimited workspaces"),
				WantStderr:      []string{"failed to execute bash command: unlimited workspaces"},
				WantRunContents: [][]string{numW},
			},
		},
		{
			name: "fails if runInt fails when getting the current workspace",
			etc: &command.ExecuteTestCase{
				Args:            []string{"left"},
				WantErr:         fmt.Errorf("failed to execute bash command: unknown workspace"),
				WantStderr:      []string{"failed to execute bash command: unknown workspace"},
				WantRunContents: [][]string{numW, cw},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces": command.IntValue(1),
					},
				},
				RunResponses: []*command.FakeRun{nRun(1), errRun("unknown workspace")},
			},
		},
		{
			name: "moves left",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(2), mcRun("DP-1")},
				Args:         []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 1",
						"xrandr --output DP-1 --brightness 1.00",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(2),
					},
				},
			},
			want: &Workspace{
				Prev: 2,
			},
		},
		{
			name: "moves left from 0 to top",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(0), mcRun("DP-2")},
				Args:         []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
						"xrandr --output DP-2 --brightness 1.00",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(0),
					},
				},
			},
			want: &Workspace{
				Prev: 0,
			},
		},
		{
			name: "left move changes brightness with trimmed arguments",
			w: &Workspace{
				Brightness: map[int]int{
					1: 37,
				},
			},
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(2), mcRun("  DP-1\t", "DP-7  ", "  \t ")},
				Args:         []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 1",
						"xrandr --output DP-1 --brightness 0.37",
						"xrandr --output DP-7 --brightness 0.37",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(2),
					},
				},
			},
			want: &Workspace{
				Prev: 2,
				Brightness: map[int]int{
					1: 37,
				},
			},
		},
		{
			name: "moves right",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(1), mcRun("eDP-9")},
				Args:         []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 2",
						"xrandr --output eDP-9 --brightness 1.00",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(1),
					},
				},
			},
			want: &Workspace{
				Prev: 1,
			},
		},
		{
			name: "moves right from last workspace",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(3), mcRun("dp1", "dp2", "dp3", "dp4")},
				Args:         []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 0",
						"xrandr --output dp1 --brightness 1.00",
						"xrandr --output dp2 --brightness 1.00",
						"xrandr --output dp3 --brightness 1.00",
						"xrandr --output dp4 --brightness 1.00",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(3),
					},
				},
			},
			want: &Workspace{
				Prev: 3,
			},
		},
		{
			name: "right move changes brightness",
			w: &Workspace{
				Brightness: map[int]int{
					0: 101,
				},
			},
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(3), mcRun("DP-1", "DP-7")},
				Args:         []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 0",
						"xrandr --output DP-1 --brightness 1.01",
						"xrandr --output DP-7 --brightness 1.01",
					},
				},
				WantRunContents: [][]string{numW, cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"numWorkspaces":    command.IntValue(4),
						"currentWorkspace": command.IntValue(3),
					},
				},
			},
			want: &Workspace{
				Prev: 3,
				Brightness: map[int]int{
					0: 101,
				},
			},
		},
		{
			name: "moves to nth workspace",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(5), mcRun()},
				Args:         []string{"3"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:       command.IntValue(3),
						"currentWorkspace": command.IntValue(5),
					},
				},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
				WantRunContents: [][]string{cw, lmCmd},
			},
			want: &Workspace{
				Prev: 5,
			},
		},
		{
			name: "nth move changes brightness",
			w: &Workspace{
				Brightness: map[int]int{
					3: 21,
				},
			},
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(5), mcRun("DP-2", "DP-5")},
				Args:         []string{"3"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:       command.IntValue(3),
						"currentWorkspace": command.IntValue(5),
					},
				},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
						"xrandr --output DP-2 --brightness 0.21",
						"xrandr --output DP-5 --brightness 0.21",
					},
				},
				WantRunContents: [][]string{cw, lmCmd},
			},
			want: &Workspace{
				Prev: 5,
				Brightness: map[int]int{
					3: 21,
				},
			},
		},
		{
			name: "does nothing if request to move to same workspace",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(2)},
				Args:         []string{"2"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:       command.IntValue(2),
						"currentWorkspace": command.IntValue(2),
					},
				},
				WantRunContents: [][]string{cw},
			},
		},
		{
			name: "moves back a workspace",
			w: &Workspace{
				Prev: 3,
			},
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(5), mcRun("dp0")},
				Args:         []string{"back"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
						"xrandr --output dp0 --brightness 1.00",
					},
				},
				WantRunContents: [][]string{cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"currentWorkspace": command.IntValue(5),
					},
				},
			},
			want: &Workspace{
				Prev: 5,
			},
		},
		{
			name: "moves back a workspace changes brightness",
			w: &Workspace{
				Prev: 3,
				Brightness: map[int]int{
					3: 45,
				},
			},
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(5), mcRun("eDP-3")},
				Args:         []string{"back"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
						"xrandr --output eDP-3 --brightness 0.45",
					},
				},
				WantRunContents: [][]string{cw, lmCmd},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"currentWorkspace": command.IntValue(5),
					},
				},
			},
			want: &Workspace{
				Prev: 5,
				Brightness: map[int]int{
					3: 45,
				},
			},
		},
		// List monitors
		{
			name: "Lists monitors",
			etc: &command.ExecuteTestCase{
				RunResponses:    []*command.FakeRun{mcRun("eDP-1", "DP-1-3")},
				Args:            []string{"monitors", "list"},
				WantRunContents: [][]string{lmCmd},
				WantStdout:      []string{"DP-1-3", "eDP-1"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"mcs": command.StringListValue("DP-1-3", "eDP-1"),
					},
				},
			},
		},
		// Set brightness
		{
			name: "Adds brightness to nil map",
			etc: &command.ExecuteTestCase{
				Args: []string{"brightness", "set", "3", "75"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:  command.IntValue(3),
						brightnessArg: command.IntValue(75),
					},
				},
			},
			want: &Workspace{
				Brightness: map[int]int{
					3: 75,
				},
			},
		},
		{
			name: "Adds brightness",
			w: &Workspace{
				Brightness: map[int]int{
					3: 75,
				},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"brightness", "set", "8", "222"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:  command.IntValue(8),
						brightnessArg: command.IntValue(222),
					},
				},
			},
			want: &Workspace{
				Brightness: map[int]int{
					3: 75,
					8: 222,
				},
			},
		},
		{
			name: "Overwrites brightness",
			w: &Workspace{
				Brightness: map[int]int{
					3: 75,
					8: 222,
				},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"brightness", "set", "8", "90"},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						workspaceArg:  command.IntValue(8),
						brightnessArg: command.IntValue(90),
					},
				},
			},
			want: &Workspace{
				Brightness: map[int]int{
					3: 75,
					8: 90,
				},
			},
		},
		{
			name: "Lists brightness",
			w: &Workspace{
				Brightness: map[int]int{
					3:  75,
					8:  222,
					24: 68,
				},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"brightness", "list"},
				WantStdout: []string{
					" 3: 75",
					" 8: 222",
					"24: 68",
				},
			},
		},
		/* Useful for commenting out tests. */
	} {
		t.Run(test.name, func(t *testing.T) {
			w := test.w
			if w == nil {
				w = CLI()
			}
			test.etc.Node = w.Node()
			command.ExecuteTest(t, test.etc)

			want := test.want
			if want == nil {
				want = &Workspace{}
			}
			command.ChangeTest(t, test.want, w, cmpopts.IgnoreUnexported(Workspace{}))
		})
	}
}

/*func TestUsage(t *testing.T) {
	command.UsageTest(t, &command.UsageTestCase{
		Node: CLI().Node(),
		WantString: []string{
			"< WORKSPACE",
			"",
			"  back",
			"",
			"  left",
			"",
			"  right",
			"",
			"Arguments:",
			"  WORKSPACE: Workspace number",
			"",
			"Symbols:",
			command.BranchDesc,
		},
	})
}
*/
