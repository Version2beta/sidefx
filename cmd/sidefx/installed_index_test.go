package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	sidefxskill "github.com/Version2beta/sidefx/skills/sidefx"
)

func TestInstalledIndexOutsideCheckout(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	binary := filepath.Join(bin, "sidefx"+suffix)
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(binary, data, 0755); err != nil {
		t.Fatal(err)
	}
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	goPath, err = filepath.Abs(goPath)
	if err != nil {
		t.Fatal(err)
	}
	goCopy := filepath.Join(bin, "go"+suffix)
	if err = os.Symlink(goPath, goCopy); err != nil {
		t.Skipf("isolated Go PATH symlink unavailable: %v", err)
	}
	home := filepath.Join(root, "home")
	dir := filepath.Join(root, "app")
	mustWrite(t, filepath.Join(dir, "go.mod"), "module example.test/installed\n\ngo 1.26\n\nrequire example.test/dependency v0.0.0\nreplace example.test/dependency => ../dependency\n")
	mustWrite(t, filepath.Join(root, "dependency", "go.mod"), "module example.test/dependency\n\ngo 1.26\n")
	mustWrite(t, filepath.Join(root, "dependency", "dep.go"), "package dependency\nimport \"os\"\nfunc Emit(){_ = os.WriteFile(\"never\",nil,0600)}\n")
	source := `package installed
import ("os";"time";"sync";"io";"log";"example.test/dependency")
func Many(){
`
	for i := 0; i < 30; i++ {
		source += fmt.Sprintf("_ = os.WriteFile(\"never-%d\",nil,0600)\n", i)
	}
	source += `}
func Families(f *os.File,m *sync.Mutex,r io.Reader){
 _,_=os.ReadFile("never");_=os.Mkdir("never",0700);_=os.Remove("never");_=os.Rename("never","never2")
 _=os.Getenv("NEVER");_=os.Setenv("NEVER","never");_=time.Now();time.Sleep(0);m.Lock();m.Unlock()
 _,_=io.ReadFull(r,nil);log.Print("never");_=f.Sync();_,_=os.Open("never");_=f.Close()
}
func Panics(){panic("never")}
func Dep(){dependency.Emit()}
func Possible(flag bool){f:=time.After;if flag{f=time.Tick};_=f(time.Second)}
func Unresolved(f func()){f()}
func Contract(r io.Reader){_,_=r.Read(nil)}
`
	mustWrite(t, filepath.Join(dir, "fixture.go"), source)
	before := snapshot(t, dir)
	// Child PATH contains exactly the copied installed binary and Go, no Python,
	// shell utilities, repository helper or checkout-relative command.
	env := append(os.Environ(), "PATH="+bin, "HOME="+home, "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOTELEMETRY=off")
	call := func(args ...string) ([]byte, []byte, int) {
		t.Helper()
		cmd := exec.Command(binary, args...)
		cmd.Dir = dir
		cmd.Env = env
		var out, diag bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &diag
		err := cmd.Run()
		if err == nil {
			return out.Bytes(), diag.Bytes(), 0
		}
		if e, ok := err.(*exec.ExitError); ok {
			return out.Bytes(), diag.Bytes(), e.ExitCode()
		}
		t.Fatal(err)
		return nil, nil, -1
	}
	success := func(args ...string) []byte {
		t.Helper()
		out, diag, status := call(args...)
		if status != 0 {
			t.Fatalf("installed %v: exit %d %s", args, status, diag)
		}
		return out
	}
	// These command shapes are the installed SKILL/reference commands, executed
	// against a real isolated executable, not merely found in markdown strings.
	success("-help")
	success("index", "-help")
	success("acquire", "-help")
	runDir := filepath.Join(root, "evidence")
	success("acquire", "-out", runDir, "-max-depth", "2", "-max-sites", "100000", "-max-edges", "10000", "./...")
	report, provenance := filepath.Join(runDir, "effects.json"), filepath.Join(runDir, "acquisition.json")
	var index map[string]any
	if err = json.Unmarshal(success("index", "-provenance", provenance, report), &index); err != nil {
		t.Fatal(err)
	}
	if index["bounds"].(map[string]any)["declared_query_categories"] != "complete" || index["category_count"].(float64) <= 10 || index["multi_category_sites"].(float64) == 0 {
		t.Fatal(index)
	}
	fromDisk, err := os.ReadFile(filepath.Join(runDir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var disk any
	json.Unmarshal(fromDisk, &disk)
	if !reflect.DeepEqual(index, disk) {
		t.Fatal("saved index differs")
	}
	preserved := snapshot(t, runDir)
	if _, _, status := call("acquire", "-out", runDir, "./nonexistent"); status == 0 || !reflect.DeepEqual(preserved, snapshot(t, runDir)) {
		t.Fatal("collision overwrote prior acquisition")
	}
	// Save separate authorized contract/unresolved views using the unchanged
	// normal CLI while Go is still available. These are distinct scopes, not
	// silently merged into the recognized acquisition.
	contractRaw := success("-contracts", "-json", "-max-depth", "2", "-max-sites", "100000", "-max-edges", "10000", "./...")
	contractPath := filepath.Join(root, "contracts.json")
	mustWrite(t, contractPath, string(contractRaw))
	unresolvedRaw := success("-unresolved", "-json", "-max-depth", "0", "-max-sites", "100000", "./...")
	unresolvedPath := filepath.Join(root, "unresolved.json")
	mustWrite(t, unresolvedPath, string(unresolvedRaw))
	// Install needs no Go. Remove even Go from the restricted PATH, leaving only
	// sidefx; absolute-path invocation continues to work even when PATH is empty.
	if err = os.Remove(goCopy); err != nil {
		t.Fatal(err)
	}
	success("skill", "install")
	success("skill", "install")
	skillDir := filepath.Join(home, ".agents", "skills", "sidefx")
	files := map[string][]byte{"SKILL.md": sidefxskill.Content}
	for k, v := range sidefxskill.Resources {
		files[k] = v
	}
	link := regexp.MustCompile(`\]\(([^)#]+)(?:#[^)]*)?\)`)
	for name, want := range files {
		path := filepath.Join(skillDir, filepath.FromSlash(name))
		actual, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(actual, want) {
			t.Fatalf("installed resource %s differs: %v", name, err)
		}
		for _, match := range link.FindAllSubmatch(actual, -1) {
			target := string(match[1])
			if strings.Contains(target, "://") {
				continue
			}
			resolved := filepath.Join(filepath.Dir(path), target)
			rel, err := filepath.Rel(skillDir, resolved)
			if err != nil || !filepath.IsLocal(rel) {
				t.Fatal("installed link escapes package", name, target)
			}
			if _, err := os.Stat(resolved); err != nil {
				t.Fatal("missing installed reference", name, target)
			}
		}
	}
	for _, name := range []string{"operation-index.md", "operation-coverage-audit.md", "operation-recognition-decisions.md"} {
		source, err := os.ReadFile(filepath.Join("..", "..", "docs", name))
		if err != nil || !bytes.Equal(source, files["references/"+name]) {
			t.Fatal("embedded/source byte mismatch", name, err)
		}
	}
	// Full index + every category/site set, entirely without Go/Python on PATH.
	var previous string
	singleton, large, possible, dependency := false, false, false, false
	for _, value := range index["categories"].([]any) {
		row := value.(map[string]any)
		category := row["category"].(string)
		if category <= previous {
			t.Fatal("category order")
		}
		previous = category
		singleton = singleton || row["sites"].(float64) == 1
		large = large || row["sites"].(float64) > 24
		possible = possible || row["possible_only"].(float64) > 0
		dependency = dependency || row["dependency_origin"].(float64) > 0
		token := ""
		ids := []string{}
		for {
			args := []string{"index", "-category", category, "-limit", "8"}
			if token != "" {
				args = append(args, "-cursor", token)
			}
			var page map[string]any
			json.Unmarshal(success(append(args, report)...), &page)
			for _, s := range page["sites"].([]any) {
				ids = append(ids, s.(map[string]any)["id"].(string))
			}
			if page["next_cursor"] == nil {
				if len(ids) != int(row["sites"].(float64)) {
					t.Fatal("incomplete category", category)
				}
				break
			}
			token = page["next_cursor"].(string)
			if _, out, status := call("index", "-category", "wrong.selector", "-cursor", token, report); status == 0 || len(out) == 0 {
				t.Fatal("wrong selector accepted")
			}
		}
		if !slices.IsSorted(ids) || len(slices.Compact(slices.Clone(ids))) != len(ids) {
			t.Fatal("unstable or duplicate sites")
		}
	}
	if !singleton || !large || !possible || !dependency {
		t.Fatal("fixture failed acceptance shape")
	}
	raw, _ := os.ReadFile(report)
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))
	if hash != index["input_sha256"] {
		t.Fatal("input identity")
	}
	var saved map[string]any
	json.Unmarshal(raw, &saved)
	sid := saved["sites"].([]any)[0].(map[string]any)["id"].(string)
	success("index", "-site", sid, report)
	success("index", "-category", "fs.write", "-limit", "8", "-offset", "8", "-expect-sha256", hash, report)
	for _, args := range [][]string{{"index", "-expect-sha256", "bad", report}, {"index", "-expect-sha256", "", report}, {"index", "-provenance", "", report}, {"index", "-category", "fs.write", "-cursor", "", report}, {"index", "-category", "fs.write", "-offset", "999999", report}, {"index", "-category", "fs.write", "-site", sid, report}, {"index", "-limit", "1", report}} {
		if out, _, status := call(args...); status == 0 || len(out) != 0 {
			t.Fatal("invalid request accepted", args)
		}
	}
	// Unknown saved fields survive retrieval; absent matching origin evidence
	// stays unknown. This deliberately authored legacy record is NOT paired with
	// the original acquisition provenance or declared unfiltered-complete.
	var enriched map[string]any
	json.Unmarshal(contractRaw, &enriched)
	contracts := enriched["interface_contracts"].([]any)
	if len(contracts) == 0 {
		t.Fatal("no contract fixture")
	}
	contract := contracts[0].(map[string]any)
	contract["future_contract"] = map[string]any{"retained": true}
	contractSID := contract["site"].(string)
	for _, value := range enriched["sites"].([]any) {
		site := value.(map[string]any)
		if site["id"] == contractSID {
			site["future_site"] = "retained"
		}
	}
	for _, value := range enriched["reverse"].([]any) {
		reverse := value.(map[string]any)
		if reverse["site"] == contractSID {
			reverse["origins"] = []any{}
		}
	}
	encoded, _ := json.Marshal(enriched)
	unknownPath := filepath.Join(root, "unknown-legacy.json")
	mustWrite(t, unknownPath, string(encoded))
	var page map[string]any
	json.Unmarshal(success("index", "-site", contractSID, unknownPath), &page)
	if page["sites"].([]any)[0].(map[string]any)["future_site"] != "retained" || !reflect.DeepEqual(page["interface_contracts"].([]any)[0], contract) {
		t.Fatal("unknown contract/site fields lost")
	}
	var unknownIndex map[string]any
	json.Unmarshal(success("index", unknownPath), &unknownIndex)
	unknown := false
	for _, value := range unknownIndex["categories"].([]any) {
		if value.(map[string]any)["unknown_origin"].(float64) > 0 {
			unknown = true
		}
	}
	if !unknown || unknownIndex["bounds"].(map[string]any)["declared_query_categories"] != "not_established" {
		t.Fatal("legacy unknown upgraded")
	}
	success("index", "-site", contractSID, contractPath)
	var unresolvedIndex map[string]any
	json.Unmarshal(success("index", unresolvedPath), &unresolvedIndex)
	if unresolvedIndex["bounds"].(map[string]any)["uncategorized_unresolved_records"].(float64) < 1 {
		t.Fatal("unresolved fixture missing")
	}
	success("index", "-uncategorized-unresolved", unresolvedPath)
	// Byte changes invalidate an input-bound continuation even when all records
	// are unchanged; the original saved report is never overwritten.
	var first map[string]any
	json.Unmarshal(success("index", "-category", "fs.write", "-limit", "8", report), &first)
	changedPath := filepath.Join(root, "changed.json")
	mustWrite(t, changedPath, string(raw)+"\n")
	if out, _, status := call("index", "-category", "fs.write", "-cursor", first["next_cursor"].(string), changedPath); status == 0 || len(out) != 0 {
		t.Fatal("stale cursor accepted")
	}
	// Missing executable on PATH is a shell-level failure, not an analyzer error.
	// Explicit absolute path is the documented alternative and works with PATH empty.
	if runtime.GOOS != "windows" {
		shell := exec.Command("/bin/sh", "-c", "command -v sidefx")
		shell.Env = append(os.Environ(), "PATH="+filepath.Join(root, "empty"))
		if err = shell.Run(); err == nil {
			t.Fatal("missing PATH binary unexpectedly found")
		}
	}
	env = append(env, "PATH="+filepath.Join(root, "empty"))
	success("index", report)
	success("skill", "install")
	if !reflect.DeepEqual(before, snapshot(t, dir)) {
		t.Fatal("fresh analysis changed source")
	}
}

func TestInstalledIndexAcquireFailureAndLegacyJSON(t *testing.T) {
	dir := copyFixture(t)
	runDir := filepath.Join(t.TempDir(), "capped")
	_, diag, status := invoke(t, dir, nil, "acquire", "-out", runDir, "-all", "-max-sites", "1", "./...")
	if status != 1 || !strings.Contains(diag, "incomplete") {
		t.Fatal(status, diag)
	}
	raw, err := os.ReadFile(filepath.Join(runDir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"declared_query_categories":"not_established"`)) {
		t.Fatal("capped orientation upgraded")
	}
	// Existing normal JSON remains the exact same pipeline and bytes under acquire.
	complete := filepath.Join(t.TempDir(), "complete")
	if _, diag, status := invoke(t, dir, nil, "acquire", "-out", complete, "-direct", "-max-depth", "0", "-max-sites", "100000", "-max-edges", "1", "./..."); status != 0 {
		t.Fatal(diag)
	}
	normal, diag, status := invoke(t, dir, nil, "-view", "effects", "-json", "-stats", "-max-sites", "100000", "-max-edges", "1", "-max-depth", "0", "-direct", "./...")
	if status != 0 {
		t.Fatal(diag)
	}
	acquired, _ := os.ReadFile(filepath.Join(complete, "effects.json"))
	if string(acquired) != normal {
		t.Fatal("ordinary analysis JSON changed")
	}
	out, diag, status := invoke(t, dir, []string{"PATH=" + t.TempDir()}, "index", filepath.Join(complete, "effects.json"))
	if status != 0 || !strings.Contains(out, `"declared_query_categories":"not_established"`) {
		t.Fatal(out, diag, status)
	}
	empty := t.TempDir()
	emptyBefore := snapshot(t, empty)
	if _, diag, status := invoke(t, dir, nil, "acquire", "-out", empty, "./missing"); status != 1 || !strings.Contains(diag, "create new run directory") || !reflect.DeepEqual(emptyBefore, snapshot(t, empty)) {
		t.Fatal("existing empty directory not preserved", diag, status)
	}
	failed := filepath.Join(t.TempDir(), "failure")
	_, _, status = invoke(t, dir, nil, "acquire", "-out", failed, "./missing")
	if status != 1 {
		t.Fatal(status)
	}
	if _, err = os.Stat(filepath.Join(failed, "acquisition.json")); err != nil {
		t.Fatal("failure provenance missing", err)
	}
	if _, err = os.Stat(filepath.Join(failed, "index.json")); !os.IsNotExist(err) {
		t.Fatal("failed analysis indexed")
	}
}
