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
		{
			name: "clean-all default all",
			argv: []string{"--clean-all"},
			want: Flags{CleanAll: true, StatusFilter: "all"},
		},
		{
			name: "clean-all pushed",
			argv: []string{"--clean-all", "--status=pushed"},
			want: Flags{CleanAll: true, StatusFilter: "pushed"},
		},
		{
			name:    "clean-all dirty without force",
			argv:    []string{"--clean-all", "--status=dirty"},
			wantErr: true,
		},
		{
			name: "clean-all dirty with force",
			argv: []string{"--clean-all", "--status=dirty", "--force"},
			want: Flags{CleanAll: true, StatusFilter: "dirty", Force: true},
		},
		{
			name:    "clean-all invalid status",
			argv:    []string{"--clean-all", "--status=foo"},
			wantErr: true,
		},
		{
			name: "clean-all dry-run yes",
			argv: []string{"--clean-all", "--dry-run", "-y"},
			want: Flags{CleanAll: true, StatusFilter: "all", DryRun: true, AssumeYes: true},
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

func TestParse_ListShortFlag(t *testing.T) {
	f, err := Parse([]string{"-L"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !f.List {
		t.Error("List = false, want true")
	}
}

func TestParse_ListLongFlag(t *testing.T) {
	f, _ := Parse([]string{"--list"})
	if !f.List {
		t.Error("List = false on --list")
	}
}

func TestParse_ListWithJSON(t *testing.T) {
	f, _ := Parse([]string{"-L", "--json"})
	if !f.List || !f.JSON {
		t.Errorf("List=%v JSON=%v", f.List, f.JSON)
	}
}

func TestParse_ListWithDir(t *testing.T) {
	f, _ := Parse([]string{"-L", "-d", "/tmp/repo"})
	if f.TargetDir != "/tmp/repo" {
		t.Errorf("TargetDir = %q", f.TargetDir)
	}
}

func TestParse_ListWithNoPRAndNoSession(t *testing.T) {
	f, _ := Parse([]string{"-L", "--no-pr", "--no-session"})
	if !f.NoPR || !f.NoSession {
		t.Errorf("NoPR=%v NoSession=%v", f.NoPR, f.NoSession)
	}
}

func TestParse_DirWithoutListErrors(t *testing.T) {
	if _, err := Parse([]string{"-d", "/x"}); err == nil {
		t.Fatal("want error: -d without -L")
	}
}

func TestParse_ListWithNewIsExclusive(t *testing.T) {
	if _, err := Parse([]string{"-L", "-n"}); err == nil {
		t.Fatal("want error: -L with -n")
	}
}

func TestParse_ListWithSuperpowersIsExclusive(t *testing.T) {
	if _, err := Parse([]string{"-L", "-s"}); err == nil {
		t.Fatal("want error: -L with -s")
	}
}

func TestParse_ListWithCleanAllIsExclusive(t *testing.T) {
	if _, err := Parse([]string{"-L", "--clean-all"}); err == nil {
		t.Fatal("want error: -L with --clean-all")
	}
}

func TestParse_ListWithPassthroughIsExclusive(t *testing.T) {
	if _, err := Parse([]string{"-L", "--", "--model", "x"}); err == nil {
		t.Fatal("want error: -L with -- passthrough")
	}
}

func TestParse_LangFlag(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"default empty", []string{}, "", false},
		{"--lang=en", []string{"--lang=en"}, "en", false},
		{"--lang=ja", []string{"--lang=ja"}, "ja", false},
		{"--lang ja", []string{"--lang", "ja"}, "ja", false},
		{"--lang invalid still parses (Init validates)", []string{"--lang=fr"}, "fr", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if got.Lang != tc.want {
				t.Errorf("Lang = %q, want %q", got.Lang, tc.want)
			}
		})
	}
}
