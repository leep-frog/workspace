package workspace

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/leep-frog/command"
)

type runIntResponse struct {
	i   int
	err error
}

func TestWorkspace(t *testing.T) {
	for _, test := range []struct {
		name     string
		w        *Workspace
		etc      *command.ExecuteTestCase
		rir      []*runIntResponse
		wantRuns [][]string
		want     *Workspace
	}{
		{
			name: "requires argument",
			etc: &command.ExecuteTestCase{
				WantErr:    fmt.Errorf("not enough arguments"),
				WantStderr: []string{"not enough arguments"},
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
			name: "requires valid argument",
			etc: &command.ExecuteTestCase{
				Args:       []string{"up"},
				WantErr:    fmt.Errorf(`strconv.Atoi: parsing "up": invalid syntax`),
				WantStderr: []string{`strconv.Atoi: parsing "up": invalid syntax`},
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
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 2,
			},
		},
		{
			name: "moves left from 0 to top",
			rir: []*runIntResponse{
				{i: 4},
				{i: 0},
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
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 0,
			},
		},
		{
			name: "moves right",
			rir: []*runIntResponse{
				{i: 4},
				{i: 1},
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
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 1,
			},
		},
		{
			name: "moves right from last workspace",
			rir: []*runIntResponse{
				{i: 4},
				{i: 3},
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
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 3,
			},
		},
		{
			name: "moves to nth workspace",
			rir: []*runIntResponse{
				{i: 5},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"3"},
				WantData: &command.Data{
					workspaceArg: command.IntValue(3),
				},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
			},
			wantRuns: [][]string{
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 5,
			},
		},
		{
			name: "moves back a workspace",
			w: &Workspace{
				Prev: 3,
			},
			rir: []*runIntResponse{
				{i: 5},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"back"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"wmctrl -s 3",
					},
				},
			},
			wantRuns: [][]string{
				{`wmctrl -d | awk '{ if ($2 == "'*'") print $1 }'`},
			},
			want: &Workspace{
				Prev: 5,
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
			w := test.w
			if w == nil {
				w = CLI()
			}
			test.etc.Node = w.Node()
			command.ExecuteTest(t, test.etc, nil)

			want := test.want
			if want == nil {
				want = &Workspace{}
			}
			command.ChangeTest(t, test.want, w, cmpopts.IgnoreUnexported(Workspace{}))

			if diff := cmp.Diff(test.wantRuns, gotContents); diff != "" {
				t.Errorf("Unexpected RunInt contents provided (-want, +got):\n%s", diff)
			}
		})
	}
}
