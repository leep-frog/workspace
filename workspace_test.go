package workspace

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/leep-frog/command"
)

type runIntResponse struct {
	i   int
	err error
}

func TestWorkspace(t *testing.T) {
	for _, test := range []struct {
		name     string
		etc      *command.ExecuteTestCase
		rir      []*runIntResponse
		wantRuns [][]string
	}{
		{
			name: "requires argument",
			etc: &command.ExecuteTestCase{
				WantErr:    fmt.Errorf("branching argument required"),
				WantStderr: []string{"branching argument required"},
			},
		},
		{
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:       []string{"up"},
				WantErr:    fmt.Errorf("argument must be one of [left right]"),
				WantStderr: []string{"argument must be one of [left right]"},
			},
		},
		{
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:       []string{"up"},
				WantErr:    fmt.Errorf("argument must be one of [left right]"),
				WantStderr: []string{"argument must be one of [left right]"},
			},
		},
		{
			name: "fails if runInt fails when getting the number of workspaces",
			rir: []*runIntResponse{
				{
					err: fmt.Errorf("unlimited workspaces"),
				},
			},
			etc: &command.ExecuteTestCase{
				Args:       []string{"left"},
				WantErr:    fmt.Errorf("failed to get number of workspaces: unlimited workspaces"),
				WantStderr: []string{"failed to get number of workspaces: unlimited workspaces"},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
			},
		},
		{
			name: "fails if runInt fails when getting the current workspace",
			rir: []*runIntResponse{
				{},
				{
					err: fmt.Errorf("unknown workspace"),
				},
			},
			etc: &command.ExecuteTestCase{
				Args:       []string{"left"},
				WantErr:    fmt.Errorf("failed to get current workspace: unknown workspace"),
				WantStderr: []string{"failed to get current workspace: unknown workspace"},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
		},
		{
			name: "moves left",
			rir: []*runIntResponse{
				{i: 4},
				{i: 2},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 1",
					},
				},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
		},
		{
			name: "moves left from 0 to top",
			rir: []*runIntResponse{
				{i: 4},
				{i: 0},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"left"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
		},
		{
			name: "moves right",
			rir: []*runIntResponse{
				{i: 4},
				{i: 1},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 2",
					},
				},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
		},
		{
			name: "moves right from last workspace",
			rir: []*runIntResponse{
				{i: 4},
				{i: 3},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"right"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 0",
					},
				},
			},
			wantRuns: [][]string{
				{"wmctrl -d | wc | awk '{ print $1 }'"},
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var gotContents [][]string
			oldRunInt := runInt
			runInt = func(contents []string) (int, error, int) {
				gotContents = append(gotContents, contents)
				if len(test.rir) == 0 {
					t.Fatalf("ran out of stubbed RunInt responses")
				}
				r := test.rir[0]
				test.rir = test.rir[1:]
				return r.i, r.err, 0
			}
			defer func() { runInt = oldRunInt }()
			test.etc.Node = CLI().Node()
			command.ExecuteTest(t, test.etc, nil)

			if diff := cmp.Diff(test.wantRuns, gotContents); diff != "" {
				t.Errorf("Unexpected RunInt contents provided (-want, +got):\n%s", diff)
			}
		})
	}
}
