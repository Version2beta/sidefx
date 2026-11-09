package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

func TestPrivateWorkspaceRebasesPaths(t *testing.T) {
	dir := t.TempDir()
	module := filepath.Join(dir, "module")
	if err := os.Mkdir(module, 0700); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(dir, "go.work")
	if err := os.WriteFile(work, []byte("go 1.24.0\nuse ./module\nreplace example.com/dep => ./dep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOWORK", work)
	cfg, cleanup, err := loadConfig(Options{Dir: module})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var path string
	for _, entry := range cfg.Env {
		if strings.HasPrefix(entry, "GOWORK=") {
			path = strings.TrimPrefix(entry, "GOWORK=")
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := modfile.ParseWork(path, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	physicalModule, err := filepath.EvalSymlinks(module)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Dir != physicalModule || len(parsed.Use) != 1 || parsed.Use[0].Path != physicalModule || len(parsed.Replace) != 1 || parsed.Replace[0].New.Path != filepath.Join(dir, "dep") {
		t.Fatalf("incorrect rebasing:\n%s", data)
	}
}

func TestPrivateWorkspaceLoadsAliasedWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	physical := filepath.Join(root, "physical")
	for _, name := range []string{"one", "two"} {
		dir := filepath.Join(physical, name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/"+name+"\n\ngo 1.26.0\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "p.go"), []byte("package "+name+"\nfunc Pure() {}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(physical, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	work := filepath.Join(alias, "go.work")
	for _, absoluteUse := range []bool{false, true} {
		uses := "use (\n./one\n./two\n)\n"
		if absoluteUse {
			uses = fmt.Sprintf("use (\n%q\n%q\n)\n", filepath.Join(alias, "one"), filepath.Join(alias, "two"))
		}
		if err := os.WriteFile(work, []byte("go 1.26.0\n"+uses), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("GOWORK", work)
		for _, tc := range []struct {
			name       string
			dir        string
			processCwd bool
			patterns   []string
			want       []string
		}{
			{"aliased-root", alias, false, []string{"./one/...", "./two/..."}, []string{"example.com/one", "example.com/two"}},
			{"aliased-module", filepath.Join(alias, "one"), false, []string{"./..."}, []string{"example.com/one"}},
			{"process-cwd", filepath.Join(alias, "one"), true, []string{"./..."}, []string{"example.com/one"}},
			{"physical-root", physical, false, []string{"./one/...", "./two/..."}, []string{"example.com/one", "example.com/two"}},
		} {
			t.Run(fmt.Sprintf("absoluteUse=%t/%s", absoluteUse, tc.name), func(t *testing.T) {
				opts := Options{Dir: tc.dir}
				if tc.processCwd {
					t.Chdir(tc.dir)
					opts.Dir = ""
				}
				cfg, cleanup, err := loadConfig(opts)
				if err != nil {
					t.Fatal(err)
				}
				defer cleanup()
				pkgs, err := packages.Load(cfg, tc.patterns...)
				if err != nil {
					t.Fatal(err)
				}
				var got []string
				for _, pkg := range pkgs {
					if len(pkg.Errors) != 0 {
						t.Fatal(pkg.Errors)
					}
					got = append(got, pkg.PkgPath)
				}
				slices.Sort(got)
				if !slices.Equal(got, tc.want) {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		}
	}
}
