package cli

import (
	"reflect"
	"testing"
)

func TestParse_Table(t *testing.T) {
	cases := []struct {
		name    string
		argv    []string
		want    Flags
		wantErr bool
	}{
		{
			name: "empty",
			argv: []string{},
			want: Flags{},
		},
		{
			name: "short help",
			argv: []string{"-h"},
			want: Flags{Help: true},
		},
		{
			name: "long help",
			argv: []string{"--help"},
			want: Flags{Help: true},
		},
		{
			name: "short version",
			argv: []string{"-v"},
			want: Flags{Version: true},
		},
		{
			name: "long version",
			argv: []string{"--version"},
			want: Flags{Version: true},
		},
		{
			name: "short new",
			argv: []string{"-n"},
			want: Flags{NewWorktree: true},
		},
		{
			name: "long new",
			argv: []string{"--new"},
			want: Flags{NewWorktree: true},
		},
		{
			name: "short superpowers implies new",
			argv: []string{"-s"},
			want: Flags{NewWorktree: true, Superpowers: true},
		},
		{
			name: "long superpowers implies new",
			argv: []string{"--superpowers"},
			want: Flags{NewWorktree: true, Superpowers: true},
		},
		{
			name: "new and superpowers combined",
			argv: []string{"-n", "-s"},
			want: Flags{NewWorktree: true, Superpowers: true},
		},
		{
			name: "double dash only",
			argv: []string{"--"},
			want: Flags{Passthrough: []string{}},
		},
		{
			name: "double dash with args",
			argv: []string{"--", "foo", "bar"},
			want: Flags{Passthrough: []string{"foo", "bar"}},
		},
		{
			name: "new with passthrough",
			argv: []string{"-n", "--", "--model", "claude-opus-4-7"},
			want: Flags{NewWorktree: true, Passthrough: []string{"--model", "claude-opus-4-7"}},
		},
		{
			name: "superpowers with passthrough",
			argv: []string{"-s", "--", "--resume"},
			want: Flags{NewWorktree: true, Superpowers: true, Passthrough: []string{"--resume"}},
		},
		{
			name:    "unknown long flag",
			argv:    []string{"--unknown"},
			wantErr: true,
		},
		{
			name:    "removed update flag",
			argv:    []string{"--update"},
			wantErr: true,
		},
		{
			name:    "removed uninstall flag",
			argv:    []string{"--uninstall"},
			wantErr: true,
		},
		{
			name:    "positional arg rejected",
			argv:    []string{"positional"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.argv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%v) expected error, got nil", tc.argv)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%v) unexpected error: %v", tc.argv, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Parse(%v):\n got  = %+v\n want = %+v", tc.argv, got, tc.want)
			}
		})
	}
}
