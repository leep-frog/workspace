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

func TestWorkspace(t *testing.T) {
	numW := []string{"set -e", "set -o pipefail", fmt.Sprintf("wmctrl -d | wc | awk '{ print $1 }'")}
	cw := []string{"set -e", "set -o pipefail", fmt.Sprintf(`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`)}

	for _, test := range []struct {
		name string
		w    *Workspace
		etc  *command.ExecuteTestCase
		want *Workspace
	}{
		{
			name: "requires argument",
			etc: &command.ExecuteTestCase{
				WantErr:         fmt.Errorf(`Argument "WORKSPACE" requires at least 1 argument, got 0`),
				WantStderr:      []string{`Argument "WORKSPACE" requires at least 1 argument, got 0`},
				WantRunContents: [][]string{cw},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"currentWorkspace": command.IntValue(1),
					},
				},
				RunResponses: []*command.FakeRun{nRun(1)},
			},
		},
		{
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:            []string{"up"},
				WantErr:         fmt.Errorf(`strconv.Atoi: parsing "up": invalid syntax`),
				WantStderr:      []string{`strconv.Atoi: parsing "up": invalid syntax`},
				WantRunContents: [][]string{cw},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"currentWorkspace": command.IntValue(2),
					},
				},
				RunResponses: []*command.FakeRun{nRun(2)},
			},
		},
		{
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:            []string{"up"},
				WantErr:         fmt.Errorf(`strconv.Atoi: parsing "up": invalid syntax`),
				WantStderr:      []string{`strconv.Atoi: parsing "up": invalid syntax`},
				WantRunContents: [][]string{cw},
				WantData: &command.Data{
					Values: map[string]*command.Value{
						"currentWorkspace": command.IntValue(3),
					},
				},
				RunResponses: []*command.FakeRun{nRun(3)},
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
				RunResponses: []*command.FakeRun{nRun(4), nRun(2)},
				Args:         []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 1",
					},
				},
				WantRunContents: [][]string{numW, cw},
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
				RunResponses: []*command.FakeRun{nRun(4), nRun(0)},
				Args:         []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
				WantRunContents: [][]string{numW, cw},
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
			name: "moves right",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(4), nRun(1)},
				Args:         []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 2",
					},
				},
				WantRunContents: [][]string{numW, cw},
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
				RunResponses: []*command.FakeRun{nRun(4), nRun(3)},
				Args:         []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 0",
					},
				},
				WantRunContents: [][]string{numW, cw},
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
			name: "moves to nth workspace",
			etc: &command.ExecuteTestCase{
				RunResponses: []*command.FakeRun{nRun(5)},
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
				WantRunContents: [][]string{cw},
			},
			want: &Workspace{
				Prev: 5,
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
				RunResponses: []*command.FakeRun{nRun(5)},
				Args:         []string{"back"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
				WantRunContents: [][]string{cw},
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

func TestUsage(t *testing.T) {
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
