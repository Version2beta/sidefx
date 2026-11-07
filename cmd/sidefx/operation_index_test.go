package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

type operationIndexRow struct {
	Category         string `json:"category"`
	Sites            int    `json:"sites"`
	Established      int    `json:"established"`
	PossibleOnly     int    `json:"possible_only"`
	Uncertain        int    `json:"uncertain"`
	Unresolved       int    `json:"unresolved"`
	SelectedOrigin   int    `json:"selected_origin"`
	DependencyOrigin int    `json:"dependency_origin"`
	UnknownOrigin    int    `json:"unknown_origin"`
}

type operationIndexOutput struct {
	Categories []operationIndexRow `json:"categories"`
	Bounds     struct {
		DeclaredQuery string `json:"declared_query_categories"`
		Uncategorized int    `json:"uncategorized_unresolved_records"`
	} `json:"bounds"`
	Sites     []analyzer.Site              `json:"sites"`
	Reverse   []analyzer.ReverseGraph      `json:"reverse"`
	Nodes     []analyzer.GraphNode         `json:"nodes"`
	Edges     []analyzer.GraphEdge         `json:"edges"`
	Contracts []analyzer.InterfaceContract `json:"interface_contracts"`
	Total     int                          `json:"total"`
	Next      *int                         `json:"next_offset"`
}

func projectOperationIndex(t *testing.T, path string, args ...string) (operationIndexOutput, string) {
	t.Helper()
	cmd := exec.Command(executable, append(append([]string{"index"}, args...), path)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native index %v: %v\n%s", args, err, out)
	}
	if os.Getenv("SIDEFX_INDEX_PYTHON_PARITY") == "1" {
		python, err := exec.LookPath("python3")
		if err != nil {
			t.Fatal("explicit development parity requires Python")
		}
		helper, err := filepath.Abs("../../tools/operation_index.py")
		if err != nil {
			t.Fatal(err)
		}
		reference := exec.Command(python, append([]string{helper, path}, args...)...)
		reference.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1")
		ref, err := reference.CombinedOutput()
		if err != nil {
			t.Fatalf("reference: %v %s", err, ref)
		}
		var nativeValue, referenceValue map[string]any
		if json.Unmarshal(out, &nativeValue) != nil || json.Unmarshal(ref, &referenceValue) != nil {
			t.Fatal("parity JSON malformed")
		}
		delete(nativeValue, "next_cursor")
		if !reflect.DeepEqual(nativeValue, referenceValue) {
			t.Fatalf("native/reference semantic mismatch: %s\n%s", out, ref)
		}
	}
	var result operationIndexOutput
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	return result, string(out)
}

func saveOperationAcquisition(t *testing.T, dir string, args ...string) (analyzer.EffectReport, string, string) {
	t.Helper()
	r, raw := effectResult(t, dir, append([]string{"-view", "effects"}, args...)...)
	out := t.TempDir()
	path, provenance := filepath.Join(out, "effects.json"), filepath.Join(out, "provenance.json")
	mustWrite(t, path, raw)
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	declaration := map[string]any{
		"input_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(raw))),
		"producer":     map[string]string{"sha256": fmt.Sprintf("%x", sha256.Sum256(binary))},
		"build":        map[string]string{"environment": "inherited offline test environment", "source": "author-known static-only fixture"},
		"cwd":          dir, "exit": 0,
		"command": append(append([]string{executable, "-json", "-view", "effects"}, args...), "./..."),
	}
	data, err := json.Marshal(declaration)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, provenance, string(data))
	return r, path, provenance
}

func TestOperationIndexCLICompleteAcquisitionAndRetrieval(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "app")
	mustWrite(t, filepath.Join(dir, "go.mod"), "module example.test/index\n\ngo 1.26\n\nrequire example.test/dependency v0.0.0\nreplace example.test/dependency => ../dependency\n")
	mustWrite(t, filepath.Join(root, "dependency", "go.mod"), "module example.test/dependency\n\ngo 1.26\n")
	mustWrite(t, filepath.Join(root, "dependency", "dep.go"), `package dependency
import "os"
func Emit() { _ = os.WriteFile("never", nil, 0600) }
`)
	source := `package index
import ("os"; "time"; "sync"; "io"; "log"; "example.test/dependency")
func Many() {
`
	for i := 0; i < 30; i++ {
		source += fmt.Sprintf("_ = os.WriteFile(\"never-%d\", nil, 0600)\n", i)
	}
	source += `}
func Families(f *os.File, m *sync.Mutex, r io.Reader) {
 _, _ = os.ReadFile("never"); _ = os.Mkdir("never", 0700); _ = os.Remove("never")
 _ = os.Rename("never", "never2"); _ = os.Getenv("NEVER"); _ = os.Setenv("NEVER", "never")
 _ = time.Now(); time.Sleep(0); m.Lock(); m.Unlock(); _, _ = io.ReadFull(r, nil)
 log.Print("never"); _ = f.Sync(); _, _ = os.Open("never"); _ = f.Close()
}
func Panics() { panic("never") }
func Dep() { dependency.Emit() }
func Possible(flag bool) { f := time.After; if flag { f = time.Tick }; _ = f(time.Second) }
func Unresolved(f func()) { f() }
`
	mustWrite(t, filepath.Join(dir, "index.go"), source)
	before := snapshot(t, root)
	r, path, provenance := saveOperationAcquisition(t, dir, "-max-depth", "2", "-max-sites", "10000", "-max-edges", "10000")
	index, text := projectOperationIndex(t, path, "--provenance", provenance)
	_, repeatedRaw := effectResult(t, dir, "-view", "effects", "-max-depth", "2", "-max-sites", "10000", "-max-edges", "10000")
	originalRaw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(originalRaw) != repeatedRaw {
		t.Fatal("repeated CLI acquisition changed bytes")
	}
	repeatedPath := filepath.Join(t.TempDir(), "repeated.json")
	mustWrite(t, repeatedPath, repeatedRaw)
	if _, repeatedIndex := projectOperationIndex(t, repeatedPath, "--provenance", provenance); repeatedIndex != text {
		t.Fatal("repeated CLI-to-index bytes changed")
	}
	if _, again := projectOperationIndex(t, path, "--provenance", provenance); text != again {
		t.Fatal("index byte ordering changed")
	}
	if len(index.Categories) <= 10 || index.Bounds.DeclaredQuery != "complete" {
		t.Fatalf("not a complete category orientation: %+v", index)
	}
	seen := map[string]bool{}
	possible := 0
	for _, row := range index.Categories {
		if seen[row.Category] {
			t.Fatalf("duplicate category %s", row.Category)
		}
		seen[row.Category] = true
		var total, established, candidate, uncertain, unresolved, selected, dependency int
		for _, site := range r.Sites {
			e, p := slices.Contains(site.Effects, row.Category), slices.Contains(site.PossibleEffects, row.Category)
			if !e && !p {
				continue
			}
			total++
			if e {
				established++
			}
			if p {
				candidate++
			}
			if site.Uncertain {
				uncertain++
				if len(site.Effects) == 0 {
					unresolved++
				}
			}
			if site.Package == "example.test/index" {
				selected++
			} else {
				dependency++
			}
		}
		want := operationIndexRow{Category: row.Category, Sites: total, Established: established, PossibleOnly: candidate, Uncertain: uncertain, Unresolved: unresolved, SelectedOrigin: selected, DependencyOrigin: dependency}
		if row != want {
			t.Fatalf("counts %s: got %+v want %+v", row.Category, row, want)
		}
		possible += row.PossibleOnly
	}
	for _, site := range r.Sites {
		for _, label := range append(slices.Clone(site.Effects), site.PossibleEffects...) {
			if !seen[label] {
				t.Fatalf("omitted category %s", label)
			}
		}
	}
	if possible == 0 {
		t.Fatal("fixture failed to exercise possible-only labels")
	}
	var write operationIndexRow
	for _, row := range index.Categories {
		if row.Category == "control.panic" && row.Sites != 1 {
			t.Fatal("singleton lost", row)
		}
		if row.Category == "fs.write" {
			write = row
		}
		if row.Category == "concurrency.timer.schedule" && (row.Established != 0 || row.PossibleOnly != 1) {
			t.Fatal("possible-only category promoted", row)
		}
	}
	if write.Sites != 32 || write.DependencyOrigin != 1 {
		t.Fatalf("large and dependency origins: %+v", write)
	}
	// Every category can be paged to completion without first-24 loss or changing IDs/evidence.
	for category := range seen {
		var retrieved []analyzer.Site
		offset := 0
		for {
			page, _ := projectOperationIndex(t, path, "--category", category, "--offset", fmt.Sprint(offset), "--limit", "7")
			nodes, edges := map[string]bool{}, map[string]analyzer.GraphEdge{}
			for _, node := range page.Nodes {
				nodes[node.ID] = true
			}
			for _, edge := range page.Edges {
				edges[edge.ID] = edge
			}
			for _, reverse := range page.Reverse {
				refs := map[string]bool{}
				for _, ref := range reverse.Nodes {
					if !nodes[ref.ID] {
						t.Fatal("page missing reverse node", ref.ID)
					}
					refs[ref.ID] = true
				}
				for _, eid := range reverse.Edges {
					edge, ok := edges[eid]
					if !ok || !refs[edge.Caller] || !refs[edge.Callee] {
						t.Fatal("page reverse edge not closed", eid)
					}
				}
			}
			retrieved = append(retrieved, page.Sites...)
			for i, site := range page.Sites {
				for _, original := range r.Sites {
					if site.ID == original.ID && !reflect.DeepEqual(site, original) {
						t.Fatal("site evidence changed", site.ID)
					}
				}
				for _, original := range r.Reverse {
					if original.Site == site.ID && !reflect.DeepEqual(page.Reverse[i], original) {
						t.Fatal("reverse evidence changed", site.ID)
					}
				}
			}
			if page.Next == nil {
				if len(retrieved) != page.Total {
					t.Fatal("pagination count mismatch")
				}
				break
			}
			offset = *page.Next
		}
		var want []string
		for _, site := range r.Sites {
			if slices.Contains(site.Effects, category) || slices.Contains(site.PossibleEffects, category) {
				want = append(want, site.ID)
			}
		}
		slices.Sort(want)
		var got []string
		for _, site := range retrieved {
			got = append(got, site.ID)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("incomplete %s retrieval", category)
		}
	}
	// Direct selected scope excludes the dependency origin, not selected category identities.
	_, directPath, directProvenance := saveOperationAcquisition(t, dir, "-direct", "-max-depth", "0", "-max-sites", "10000")
	direct, _ := projectOperationIndex(t, directPath, "--provenance", directProvenance)
	if direct.Bounds.DeclaredQuery != "complete" {
		t.Fatal(direct.Bounds)
	}
	for _, row := range direct.Categories {
		if row.DependencyOrigin != 0 {
			t.Fatal("direct scope gained dependency", row)
		}
	}
	// -all exposes function summaries but DOES NOT bypass ordinary origin output caps.
	capped, capPath, capProvenance := saveOperationAcquisition(t, dir, "-all", "-max-depth", "0", "-max-sites", "1")
	capIndex, _ := projectOperationIndex(t, capPath, "--provenance", capProvenance)
	if !capped.Truncated.Sites || len(capped.Inventory) == 0 || len(capped.Sites) != 1 || capIndex.Bounds.DeclaredQuery != "not_established" {
		t.Fatal("-all/cap falsely complete")
	}
	for _, filter := range [][]string{{"-effect", "fs.sync"}, {"-package", "index"}, {"-function", "=Many"}, {"-site", r.Sites[0].ID}} {
		_, filteredPath, filteredProvenance := saveOperationAcquisition(t, dir, append([]string{"-max-depth", "0", "-max-sites", "10000"}, filter...)...)
		filtered, _ := projectOperationIndex(t, filteredPath, "--provenance", filteredProvenance)
		if filtered.Bounds.DeclaredQuery != "not_established" {
			t.Fatalf("filtered scope called complete: %v", filter)
		}
		if filter[0] == "-effect" {
			var labels []string
			for _, row := range filtered.Categories {
				labels = append(labels, row.Category)
			}
			if !slices.Equal(labels, []string{"fs.sync", "fs.write"}) {
				t.Fatal("group projection hid co-label", labels)
			}
		}
	}
	u, up, _ := saveOperationAcquisition(t, dir, "-unresolved", "-max-depth", "0", "-max-sites", "10000")
	ui, _ := projectOperationIndex(t, up)
	uncategorized := 0
	for _, site := range u.Sites {
		if len(site.Effects) == 0 && len(site.PossibleEffects) == 0 && site.Uncertain {
			uncategorized++
		}
	}
	if uncategorized == 0 || ui.Bounds.Uncategorized != uncategorized {
		t.Fatal("unresolved records not separated")
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Fatal("analysis modified fixture")
	}
}

func TestOperationIndexCLIContractsAndCandidateCallers(t *testing.T) {
	dir := operationCoverageFixture(t)
	mustWrite(t, filepath.Join(dir, "index_candidates.go"), `package operationcoverage
import "os"
type indexBackend interface { Save() error }
type indexDisk struct{}
func (indexDisk) Save() error { return os.WriteFile("never", nil, 0600) }
func indexFactory() indexBackend { return indexDisk{} }
func IndexCandidate() error { return indexFactory().Save() }
`)
	r, path, _ := saveOperationAcquisition(t, dir, "-contracts", "-max-sites", "10000", "-max-edges", "10000", "-max-depth", "2")
	candidate := false
	for _, edge := range r.Edges {
		candidate = candidate || edge.Candidate
	}
	if !candidate || len(r.Contracts) == 0 {
		t.Fatal("fixture must exercise candidates and contracts")
	}
	page, _ := projectOperationIndex(t, path, "--category", "fs.write", "--limit", "100")
	candidatePreserved := false
	for _, edge := range page.Edges {
		for _, original := range r.Edges {
			if edge.ID == original.ID && !reflect.DeepEqual(edge, original) {
				t.Fatal("edge evidence changed")
			}
		}
		candidatePreserved = candidatePreserved || edge.Candidate
	}
	if !candidatePreserved {
		t.Fatal("candidate caller evidence lost")
	}
	for _, node := range page.Nodes {
		for _, original := range r.Nodes {
			if node.ID == original.ID && !reflect.DeepEqual(node, original) {
				t.Fatal("node evidence changed")
			}
		}
	}
	for _, contract := range r.Contracts {
		if !strings.Contains(contract.Method, "io.") {
			continue
		}
		page, _ := projectOperationIndex(t, path, "--site", contract.Site)
		if len(page.Contracts) == 0 {
			t.Fatal("contract lost", contract.Site)
		}
		for _, c := range page.Contracts {
			if c.Site == contract.Site && !reflect.DeepEqual(c, contract) {
				t.Fatal("contract changed")
			}
		}
		return
	}
	t.Fatal("no IO contract exercised")
}
