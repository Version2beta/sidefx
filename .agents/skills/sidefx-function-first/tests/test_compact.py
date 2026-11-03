"""Stdlib tests; SIDEFX_BIN opts into an isolated, author-known fixture dry run."""

import datetime
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

SKILL = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("compact", SKILL / "scripts/compact.py")
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class CompactTests(unittest.TestCase):
    def test_preserves_established_possible_and_unknowns(self):
        record = {"package": "example.test/pkg", "function": "(*Store).Save",
                  "effects": [{"identifier": "fs.write", "direct": True},
                              {"identifier": "fs.read", "direct": False}],
                  "possible_effects": ["fs.remove"], "uncertain": True}
        row, = MODULE.compact([record])
        self.assertEqual(row, {
            "package": record["package"], "function": record["function"],
            "location": None, "direct_marked": ["fs.write"],
            "inherited_marked": ["fs.read"], "possible_effects": ["fs.remove"],
            "uncertain": True, "uncertainty_reasons": None,
            "uncertainty_provenance": None,
        })
        self.assertEqual(record["effects"][0]["direct"], True)

    def test_empty_summary_is_retained_without_safety_label(self):
        row, = MODULE.compact([{"package": "p", "function": "Empty",
                                "effects": [], "uncertain": False}])
        self.assertEqual(row["direct_marked"], [])
        self.assertEqual(row["inherited_marked"], [])
        self.assertEqual(row["possible_effects"], [])
        self.assertNotIn("safe", row)

    def test_rejects_wrong_projection_and_malformed_fields(self):
        valid = {"package": "p", "function": "F", "effects": [], "uncertain": False}
        for bad in ({"records": []}, [None], [{**valid, "package": None}],
                    [{**valid, "uncertain": "false"}],
                    [{**valid, "effects": None}],
                    [{**valid, "effects": [{"identifier": "fs.read", "direct": 1}]}],
                    [{**valid, "possible_effects": "fs.write"}]):
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                MODULE.compact(bad)

    def test_cli_malformed_input_fails_without_partial_output(self):
        with tempfile.TemporaryDirectory() as root:
            source = Path(root) / "bad.json"
            source.write_text('[{"package":"p","function":"F","effects":[], '
                              '"uncertain":false},{}]', encoding="utf-8")
            result = subprocess.run([sys.executable, str(SKILL / "scripts/compact.py"),
                                     str(source)], capture_output=True, text=True)
            self.assertEqual(result.returncode, 1)
            self.assertEqual(result.stdout, "")
            self.assertIn("record 1", result.stderr)


@unittest.skipUnless(os.environ.get("SIDEFX_BIN"), "set SIDEFX_BIN for real CLI dry run")
class FunctionFirstDryRun(unittest.TestCase):
    def test_real_summary_then_prospective_selection_and_isolated_probe(self):
        binary = str(Path(os.environ["SIDEFX_BIN"]).resolve(strict=True))
        root = Path(tempfile.mkdtemp(prefix="dry-run-", dir=os.environ.get("SIDEFX_EVIDENCE")))
        print(f"\nRetained dry-run evidence: {root}", flush=True)
        target = root / "target"
        target.mkdir()
        env = dict(os.environ)
        for key in ("HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME",
                    "XDG_STATE_HOME", "TMPDIR", "GOCACHE", "GOMODCACHE"):
            directory = root / key.lower()
            directory.mkdir()
            env[key] = str(directory)
        env.update(GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off",
                   GOFLAGS="-mod=readonly", GOWORK="off")
        (root / "environment.json").write_text(json.dumps(
            {key: env[key] for key in ("HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME",
             "XDG_CACHE_HOME", "XDG_STATE_HOME", "TMPDIR", "GOCACHE", "GOMODCACHE",
             "GOTOOLCHAIN", "GOPROXY", "GOSUMDB", "GOFLAGS", "GOWORK")}, indent=2))

        def event(kind, **fields):
            with (root / "chronology.jsonl").open("a") as log:
                log.write(json.dumps({"utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
                                      "event": kind, **fields}) + "\n")

        def run(label, args):
            event("command-start", label=label, cwd=str(target), argv=args)
            result = subprocess.run(args, cwd=target, env=env, capture_output=True, timeout=120)
            (root / f"{label}.stdout").write_bytes(result.stdout)
            (root / f"{label}.stderr").write_bytes(result.stderr)
            (root / f"{label}.exit").write_text(str(result.returncode) + "\n")
            event("command-end", label=label, status=result.returncode)
            self.assertEqual(result.returncode, 0, result.stderr.decode(errors="replace"))
            return result

        # This is authored test data, not blind discovery. Its contract and injected
        # violation are known to the author before any summary selection.
        event("fixture-created", prior_knowledge="author-known injected defect; not blind discovery")
        (target / "go.mod").write_text("module example.test/dryrun\n\ngo 1.26\n")
        source = target / "fixture.go"
        source.write_text('''package dryrun
import ("errors"; "os")
func write(path string, data []byte) error { return os.WriteFile(path, data, 0600) }
// Replace must leave the destination unchanged on any error, including denial.
func Replace(path string, data []byte, allow bool) error {
    if err := write(path, data); err != nil { return err }
    if !allow { return errors.New("denied") }
    return nil
}
func read(path string) error { _, err := os.ReadFile(path); return err }
func remove(path string) error { return os.Remove(path) }
func Possible(flag bool, path string) error {
    f := read
    if flag { f = remove }
    return f(path)
}
func Unknown(f func(string) error, path string) error { return f(path) }
func Empty() {}
''')
        before = hashlib.sha256(source.read_bytes()).hexdigest()
        raw = run("functions", [binary, "-view", "functions", "-json", "-stats", "./..."])
        records = json.loads(raw.stdout)
        stats = json.loads(raw.stderr)
        (root / "binary.sha256").write_text(hashlib.sha256(Path(binary).read_bytes()).hexdigest() + "\n")
        run("go-version", ["go", "version"])
        run("binary-build", ["go", "version", "-m", binary])
        reduced = run("compact", [sys.executable, str(SKILL / "scripts/compact.py"),
                                  str(root / "functions.stdout")])
        rows = [json.loads(line) for line in reduced.stdout.splitlines()]
        by_name = {row["function"]: row for row in rows}
        self.assertEqual(len(rows), len(records))
        self.assertEqual(stats["packages_selected"], 1)
        self.assertEqual(stats["source_functions"], 7)
        self.assertEqual(stats["omitted"], 1)
        self.assertIn("fs.write", by_name["write"]["direct_marked"])
        selected = by_name["Replace"]
        self.assertFalse(selected["direct_marked"])
        self.assertIn("fs.write", selected["inherited_marked"])
        self.assertFalse(by_name["Possible"]["direct_marked"])
        self.assertFalse(by_name["Possible"]["inherited_marked"])
        self.assertEqual(set(by_name["Possible"]["possible_effects"]), {"fs.read", "fs.remove"})
        self.assertTrue(by_name["Unknown"]["uncertain"])
        self.assertIsNone(selected["location"])
        selection = {
            "utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "classification": selected,
            "location": "unresolved before implementation read",
            "obligation": "denied replacement leaves the destination unchanged (fixture contract)",
            "why": "low-direct caller owns the decision over inherited filesystem writes",
            "question": "does denial precede mutation or can an error leave changed content?",
            "prior_knowledge": "fixture author knows injected defect; procedural dry run only",
        }
        (root / "selection.json").write_text(json.dumps(selection, indent=2) + "\n")
        event("selection-saved", path="selection.json")
        event("implementation-read", path=str(source), lines="1-18")
        inspected = source.read_text()
        self.assertIn("if !allow", inspected)
        self.assertIn("func Replace(", inspected.splitlines()[4])
        event("source-observation", location="fixture.go:5", contract="fixture.go:4",
              observation="write at line 6 precedes denial at line 7")
        inventory = run("functions-all", [binary, "-view", "functions", "-all",
                                         "-json", "-stats", "./..."])
        all_rows = MODULE.compact(json.loads(inventory.stdout))
        self.assertEqual(len(all_rows), 7)
        empty, = [row for row in all_rows if row["function"] == "Empty"]
        self.assertEqual(empty["direct_marked"] + empty["inherited_marked"]
                         + empty["possible_effects"], [])
        self.assertFalse(empty["uncertain"])  # No purity inference.
        # Narrow query verifies exact function and package-substring semantics.
        focused = run("exact-function", [binary, "-view", "functions", "-json",
                      "-function", "=Replace", "-package", "test/dry", "./..."])
        self.assertEqual([r["function"] for r in json.loads(focused.stdout)], ["Replace"])
        event("probe-authorized", isolation="fresh paths; stdlib-only fixture; no app init, listener or network")
        (target / "fixture_test.go").write_text('''package dryrun
import ("os"; "path/filepath"; "testing")
func TestDeniedReplacementMutatesFixture(t *testing.T) {
    path := filepath.Join(t.TempDir(), "document")
    if err := os.WriteFile(path, []byte("old"), 0600); err != nil { t.Fatal(err) }
    err := Replace(path, []byte("new"), false)
    got, readErr := os.ReadFile(path)
    if err == nil || readErr != nil || string(got) != "new" {
        t.Fatalf("expected denied mutation: err=%v content=%q readErr=%v", err, got, readErr)
    }
}
''')
        run("probe", ["go", "test", "-run", "^TestDeniedReplacementMutatesFixture$",
                      "-count=1", "-timeout=30s", "./..."])
        self.assertEqual(hashlib.sha256(source.read_bytes()).hexdigest(), before)
        (root / "conclusions.json").write_text(json.dumps({
            "demonstrated_fixture_defect": "Replace returns denial after overwriting temp destination",
            "evidence": "probe asserts denial plus changed bytes; test exit 0",
            "untested_hypothesis": "a lower-level write error might also leave changed content; no failing-write probe",
            "limits": "author-known synthetic fixture, not target discovery, production impact or novelty",
            "source_sha256_before_and_after": before,
        }, indent=2) + "\n")


if __name__ == "__main__":
    unittest.main()
