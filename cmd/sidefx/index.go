package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Version2beta/sidefx/internal/operationindex"
)

// recordedGoProxy preserves ordinary proxy policy but never records URL userinfo
// or query parameters. Redact the whole field rather than expose secret-derived
// hashes or a partially scrubbed proxy list. Malformed URLs are also withheld.
func recordedGoProxy(value string) string {
	for _, entry := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '|' }) {
		candidate := entry
		if !strings.Contains(entry, ":/") && !strings.HasPrefix(entry, "/") {
			candidate = "//" + entry
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery {
			return "[REDACTED proxy URL credentials/query]"
		}
	}
	return value
}

func runIndex(args []string, stdout, stderr io.Writer) int {
	f := flag.NewFlagSet("sidefx index", flag.ContinueOnError)
	f.SetOutput(stderr)
	provenance := f.String("provenance", "", "hash-bound acquisition JSON (optional for legacy reports)")
	expect := f.String("expect-sha256", "", "reject changed report bytes")
	category := f.String("category", "", "retrieve an exact category")
	site := f.String("site", "", "retrieve an exact canonical site ID")
	uncategorized := f.Bool("uncategorized-unresolved", false, "retrieve truly uncategorized unresolved sites")
	offset := f.Int("offset", 0, "offset in sorted selected site IDs")
	limit := f.Int("limit", 10, "sites per page (1..100); graphs retain input bounds")
	cursor := f.String("cursor", "", "input/selector-bound continuation from next_cursor")
	f.Usage = func() {
		fmt.Fprintln(stderr, "Usage: sidefx index [flags] REPORT.json\nIndex saved schema 1 JSON without Go, Python, source loading or SSA. Flags precede the report.")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if f.NArg() != 1 {
		f.Usage()
		return 2
	}
	specified := map[string]bool{}
	f.Visit(func(f *flag.Flag) { specified[f.Name] = true })
	fail := func(err error) int { fmt.Fprintln(stderr, "sidefx index:", err); return 1 }
	raw, err := os.ReadFile(f.Arg(0))
	if err != nil {
		return fail(err)
	}
	r, err := operationindex.Load(raw)
	if err != nil {
		return fail(err)
	}
	if specified["expect-sha256"] && *expect != r.SHA256 {
		return fail(fmt.Errorf("input SHA-256 mismatch"))
	}
	var p any
	if specified["provenance"] {
		raw, err = os.ReadFile(*provenance)
		if err != nil {
			return fail(err)
		}
		p, err = operationindex.Decode(raw)
		if err != nil {
			return fail(err)
		}
		if p == nil {
			return fail(fmt.Errorf("provenance: expected object"))
		}
	}
	if specified["cursor"] && *cursor == "" {
		return fail(fmt.Errorf("invalid empty cursor"))
	}
	var result map[string]any
	selecting := specified["category"] || specified["site"] || specified["uncategorized-unresolved"]
	if selecting {
		selectors := 0
		for _, k := range []string{"category", "site", "uncategorized-unresolved"} {
			if specified[k] {
				selectors++
			}
		}
		if selectors != 1 || (*cursor != "" && specified["offset"]) {
			return fail(fmt.Errorf("use one nonempty selector; cursor and offset are mutually exclusive"))
		}
		result, err = r.Page(operationindex.Selector{Category: *category, Site: *site, Uncategorized: *uncategorized}, *offset, *limit, *cursor, p)
	} else {
		if specified["offset"] || specified["limit"] || specified["cursor"] {
			return fail(fmt.Errorf("offset/limit/cursor require a retrieval selector"))
		}
		result, err = r.Index(p)
	}
	if err != nil {
		return fail(err)
	}
	if err = json.NewEncoder(stdout).Encode(result); err != nil {
		return fail(err)
	}
	return 0
}

// acquire deliberately exposes only unfiltered recognized-operation acquisition.
// Existing normal analysis remains the interface for filters/contracts/unresolved.
func runAcquire(args []string, stdout, stderr io.Writer) int {
	f := flag.NewFlagSet("sidefx acquire", flag.ContinueOnError)
	f.SetOutput(stderr)
	out := f.String("out", "", "new nonexistent run directory under an authorized parent")
	direct := f.Bool("direct", false, "selected-root origins only (default includes reachable dependencies/candidate paths)")
	tests := f.Bool("tests", false, "include test source without executing tests")
	exact := f.Bool("exact", false, "do not expand local directory packages")
	all := f.Bool("all", false, "also retain function inventory; does not bypass caps")
	sites := f.Int("max-sites", 100000, "positive origin-output cap; completeness is checked, not assumed")
	edges := f.Int("max-edges", 1, "positive reverse-edge cap")
	depth := f.Int("max-depth", 0, "nonnegative reverse caller depth, independent of category visibility")
	depFunctions := f.Int("dependency-vta-functions", 0, "paired positive dependency VTA function budget")
	depEdges := f.Int("dependency-vta-edges", 0, "paired positive dependency VTA edge budget")
	f.Usage = func() {
		fmt.Fprintln(stderr, "Usage: sidefx acquire -out NEW_DIR [flags] PACKAGES\nSave unfiltered effects, native provenance and index exclusively. Requires Go for analysis; never executes target code.")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	specified := map[string]bool{}
	f.Visit(func(f *flag.Flag) { specified[f.Name] = true })
	if *out == "" || f.NArg() == 0 || *sites < 1 || *edges < 1 || *depth < 0 || specified["dependency-vta-functions"] != specified["dependency-vta-edges"] || (specified["dependency-vta-functions"] && (*depFunctions < 1 || *depEdges < 1)) {
		f.Usage()
		return 2
	}
	for _, p := range f.Args() {
		if len(p) == 0 || p[0] == '-' {
			f.Usage()
			return 2
		}
	}
	fail := func(err error) int { fmt.Fprintln(stderr, "sidefx acquire:", err); return 1 }
	// Refuse collisions before source loading, producer inspection, or any file writes.
	if err := os.Mkdir(*out, 0755); err != nil {
		return fail(fmt.Errorf("create new run directory: %w", err))
	}
	path, err := os.Executable()
	if err != nil {
		return fail(err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return fail(err)
	}
	binary, err := os.ReadFile(path)
	if err != nil {
		return fail(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	build := map[string]any{"runtime": runtime.Version(), "source_identity": "not captured; caller must retain target revision and dirty source identities"}
	env := map[string]string{}
	for _, k := range []string{"GOPROXY", "GONOPROXY", "GOSUMDB", "GOTOOLCHAIN", "GOFLAGS", "GOWORK", "GOTELEMETRY", "GOOS", "GOARCH", "CGO_ENABLED"} {
		env[k] = os.Getenv(k)
	}
	env["GOPROXY"] = recordedGoProxy(env["GOPROXY"])
	build["environment"] = env
	if info, ok := debug.ReadBuildInfo(); ok {
		build["binary_build_info"] = info
	}
	goenv, err := exec.Command("go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOWORK").Output()
	if err != nil {
		return fail(fmt.Errorf("Go toolchain unavailable: %w", err))
	}
	effective, err := operationindex.Decode(goenv)
	if err != nil {
		return fail(err)
	}
	build["go_env"] = effective
	command := []string{"-view", "effects", "-json", "-stats", "-max-sites", strconv.Itoa(*sites), "-max-edges", strconv.Itoa(*edges), "-max-depth", strconv.Itoa(*depth)}
	for _, b := range []struct {
		name  string
		value bool
	}{{"direct", *direct}, {"tests", *tests}, {"exact", *exact}, {"all", *all}} {
		if b.value {
			command = append(command, "-"+b.name)
		}
	}
	if *depFunctions > 0 {
		command = append(command, "-dependency-vta-functions", strconv.Itoa(*depFunctions), "-dependency-vta-edges", strconv.Itoa(*depEdges))
	}
	command = append(command, f.Args()...)
	record := map[string]any{"command": append([]string{path}, command...), "cwd": cwd, "build": build, "producer": map[string]string{"path": path, "sha256": fmt.Sprintf("%x", sha256.Sum256(binary))}}
	writeJSON := func(name string, value any) error {
		file, err := os.OpenFile(filepath.Join(*out, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		encErr := json.NewEncoder(file).Encode(value)
		closeErr := file.Close()
		if encErr != nil {
			return encErr
		}
		return closeErr
	}
	if err = writeJSON("acquisition-start.json", record); err != nil {
		return fail(err)
	}
	report, err := os.OpenFile(filepath.Join(*out, "effects.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fail(err)
	}
	diagnostic, err := os.OpenFile(filepath.Join(*out, "effects.stderr"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		report.Close()
		return fail(err)
	}
	status := run(command, report, diagnostic)
	reportErr := report.Close()
	diagnosticErr := diagnostic.Close()
	raw, err := os.ReadFile(filepath.Join(*out, "effects.json"))
	if err != nil {
		return fail(err)
	}
	record["exit"] = status
	record["input_sha256"] = fmt.Sprintf("%x", sha256.Sum256(raw))
	if err = writeJSON("acquisition.json", record); err != nil {
		return fail(err)
	}
	if reportErr != nil {
		return fail(reportErr)
	}
	if diagnosticErr != nil {
		return fail(diagnosticErr)
	}
	if status != 0 {
		return fail(fmt.Errorf("analysis failed (exit %d); retain effects.stderr, not a zero inventory", status))
	}
	r, err := operationindex.Load(raw)
	if err != nil {
		return fail(err)
	}
	// Decode our record using the same numeric representation as saved provenance.
	provenanceRaw, err := json.Marshal(record)
	if err != nil {
		return fail(err)
	}
	p, err := operationindex.Decode(provenanceRaw)
	if err != nil {
		return fail(err)
	}
	index, err := r.Index(p)
	if err != nil {
		return fail(err)
	}
	if err = writeJSON("index.json", index); err != nil {
		return fail(err)
	}
	if index["bounds"].(map[string]any)["declared_query_categories"] != "complete" {
		return fail(fmt.Errorf("site output incomplete; preserved index is not complete declared-query orientation"))
	}
	fmt.Fprintln(stdout, "Saved acquisition and complete declared-query category index:", *out)
	return 0
}
