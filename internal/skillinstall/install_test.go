package skillinstall

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstallCreatesAndUpdatesManagedSkill(t *testing.T) {
	home := t.TempDir()
	first, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("first\n")})
	if err != nil || !first.Updated {
		t.Fatalf("first install: %+v, %v", first, err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(first.Path), ".sidefx-sha256")); err != nil {
		t.Fatalf("sidefx ownership marker: %v", err)
	}
	data, err := os.ReadFile(first.Path)
	if err != nil || string(data) != "first\n" {
		t.Fatalf("installed content: %q, %v", data, err)
	}
	mode, err := os.Stat(first.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !mode.Mode().IsRegular() {
		t.Fatalf("skill is not a regular file: %v", mode.Mode())
	}
	if runtime.GOOS == "windows" {
		// Windows exposes the read-only attribute, not Unix permission bits.
		if mode.Mode().Perm()&0200 == 0 {
			t.Fatalf("skill is read-only: %v", mode.Mode())
		}
	} else if mode.Mode().Perm() != 0644 {
		t.Fatalf("skill permissions: %v", mode.Mode())
	}
	second, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("second\n")})
	if err != nil || !second.Updated {
		t.Fatalf("update: %+v, %v", second, err)
	}
	data, _ = os.ReadFile(second.Path)
	if string(data) != "second\n" {
		t.Fatalf("updated content: %q", data)
	}
	third, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("second\n")})
	if err != nil || third.Updated {
		t.Fatalf("repeat install: %+v, %v", third, err)
	}
}

func TestInstallRefusesUnmanagedOrModifiedSkill(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".agents", "skills", "sidefx", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("custom\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("canonical\n")}); err == nil {
		t.Fatal("unmanaged skill was overwritten")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "custom\n" {
		t.Fatalf("custom skill changed: %q", data)
	}
}

func TestInstallReportsWriteFailure(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".agents", "skills", "sidefx")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("skill\n")}); err == nil {
		t.Fatal("expected installation failure")
	}
}

func TestInstallResourcePackagePreflightAndUpgrade(t *testing.T) {
	home := t.TempDir()
	cfg := Config{Home: home, Name: "sidefx", Content: []byte("old skill"), Resources: map[string][]byte{"references/a.md": []byte("old reference")}}
	first, err := Install(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(first.Path)
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("user notes"), 0644); err != nil {
		t.Fatal(err)
	}
	// An unmanaged collision in a later-sorted file must prevent the known SKILL
	// upgrade and every earlier resource change.
	if err := os.WriteFile(filepath.Join(dir, "references", "z.md"), []byte("user"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg.Content = []byte("new skill")
	cfg.Resources = map[string][]byte{"references/a.md": []byte("new reference"), "references/z.md": []byte("canonical")}
	if _, err := Install(cfg); err == nil {
		t.Fatal("unmanaged resource overwritten")
	}
	for name, want := range map[string]string{"SKILL.md": "old skill", "references/a.md": "old reference", "references/z.md": "user", "notes.txt": "user notes"} {
		data, _ := os.ReadFile(filepath.Join(dir, name))
		if string(data) != want {
			t.Fatal("preflight changed", name)
		}
	}
	// Identical bytes are adopted, just like the legacy SKILL.md policy.
	if err := os.WriteFile(filepath.Join(dir, "references", "z.md"), []byte("canonical"), 0644); err != nil {
		t.Fatal(err)
	}
	if result, err := Install(cfg); err != nil || !result.Updated {
		t.Fatal(result, err)
	}
	if result, err := Install(cfg); err != nil || result.Updated {
		t.Fatal(result, err)
	}
	cfg.Content = []byte("next skill")
	cfg.Resources["references/a.md"] = []byte("next reference")
	if err := os.WriteFile(filepath.Join(dir, "references", "a.md"), []byte("user modification"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(cfg); err == nil {
		t.Fatal("modified resource overwritten")
	}
	data, _ := os.ReadFile(first.Path)
	if string(data) != "new skill" {
		t.Fatal("partial SKILL overwrite")
	}
}

func TestInstallLegacyAdoptionAndResourceState(t *testing.T) {
	home := t.TempDir()
	cfg := Config{Home: home, Name: "sidefx", Content: []byte("legacy")}
	first, err := Install(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Content = []byte("upgraded")
	cfg.Resources = map[string][]byte{"references/a.md": []byte("reference")}
	if _, err := Install(cfg); err != nil {
		t.Fatal("legacy marker upgrade", err)
	}
	dir := filepath.Dir(first.Path)
	data, _ := os.ReadFile(filepath.Join(dir, ".sidefx-resources.json"))
	if len(data) == 0 {
		t.Fatal("missing resource marker")
	}
	if err := os.WriteFile(filepath.Join(dir, ".sidefx-resources.json"), []byte(`{"../escape":"bad"}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg.Content = []byte("should not write")
	if _, err := Install(cfg); err == nil {
		t.Fatal("malformed manifest accepted")
	}
	data, _ = os.ReadFile(first.Path)
	if string(data) != "upgraded" {
		t.Fatal("malformed state caused write")
	}
	// Existing identical unmanaged SKILL and resource files can be adopted.
	other := t.TempDir()
	base := filepath.Join(other, ".agents", "skills", "sidefx")
	if err := os.MkdirAll(filepath.Join(base, "references"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(base, "SKILL.md"), []byte("same"), 0644)
	os.WriteFile(filepath.Join(base, "references", "a.md"), []byte("same"), 0644)
	if result, err := Install(Config{Home: other, Name: "sidefx", Content: []byte("same"), Resources: map[string][]byte{"references/a.md": []byte("same")}}); err != nil || result.Updated {
		t.Fatal(result, err)
	}
}

func TestInstallRejectsResourcePathsAndSymlinks(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "references/../escape", "references//a", "references/./a", "references\\a", "SKILL.md", ".sidefx-sha256", "references/.sidefx-resources.json", ""} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			if _, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("skill"), Resources: map[string][]byte{name: []byte("resource")}}); err == nil {
				t.Fatal("bad path accepted")
			}
			if _, err := os.Stat(filepath.Join(home, ".agents")); !os.IsNotExist(err) {
				t.Fatal("invalid resource created directories")
			}
		})
	}
	for _, name := range []string{"SKILL.md", "references", "references/a.md", ".sidefx-sha256", ".sidefx-resources.json", ".agents"} {
		t.Run("symlink-"+name, func(t *testing.T) {
			home := t.TempDir()
			outside := t.TempDir()
			base := filepath.Join(home, ".agents", "skills", "sidefx")
			path := filepath.Join(base, name)
			if name == ".agents" {
				path = filepath.Join(home, name)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(outside, "target")
			if name == "references" || name == ".agents" {
				os.Mkdir(target, 0755)
			} else {
				os.WriteFile(target, []byte("user"), 0644)
			}
			if err := os.Symlink(target, path); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			if _, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("skill"), Resources: map[string][]byte{"references/a.md": []byte("resource")}}); err == nil {
				t.Fatal("symlink accepted")
			}
			if name != "references" && name != ".agents" {
				data, _ := os.ReadFile(target)
				if string(data) != "user" {
					t.Fatal("symlink target changed")
				}
			}
		})
	}
	home := t.TempDir()
	if _, err := Install(Config{Home: home, Name: "../escape", Content: []byte("skill")}); err == nil {
		t.Fatal("name escape accepted")
	}
	if _, err := Install(Config{Home: home, Name: "sidefx", Content: []byte("skill"), Resources: map[string][]byte{"references": []byte("a"), "references/a.md": []byte("b")}}); err == nil {
		t.Fatal("file/dir collision accepted")
	}
}
