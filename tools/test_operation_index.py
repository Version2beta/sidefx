"""Saved-report contracts; real CLI coverage lives in operation_index_test.go."""
import copy
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from operation_index import SavedReport, declared_query


def fixture():
    sites, reverse, nodes = [], [], []
    for i, (effects, possible, uncertain) in enumerate([
            (["a", "b"], [], True), (["a"], [], False),
            ([], ["c"], False), ([], ["c"], True), ([], [], True)]):
        sid, nid = f"site:{i}", f"node:{i}"
        sites.append(dict(id=sid, package="p", function="F", kind="call", api="API",
                          apis=["API"], operations=[], location={"file": "f.go", "line": i+1, "column": 1},
                          effects=effects, possible_effects=possible, uncertain=uncertain,
                          uncertainty_reasons=["callback"] if uncertain else [], recognition="summary"))
        nodes.append(dict(id=nid, package="p", function="F", selected=i != 1))
        reverse.append(dict(site=sid, origins=[nid] if i != 2 else [],
                            nodes=[{"id": nid, "depth": 0}], edges=[], truncated=dict(sites=False, edges=False, depth=True)))
    return dict(schema_version=1, sites=sites, groups=[], reverse=reverse, nodes=nodes, edges=[],
                unresolved_sites=["site:3", "site:4"],
                coverage=dict(packages=["p"], sites_emitted=5, sites_matching_filters=5,
                              operation_sites_identified=2, selected_operation_sites=1,
                              dependency_operation_sites=1, unresolved_sites_identified=2),
                truncated=dict(sites=False, edges=False, depth=True),
                interface_contracts=[{"site": "site:0", "open": True, "implementations": []}])


def provenance(digest="hash", *flags):
    return dict(input_sha256=digest, producer={"sha256": "a" * 64}, build={"GOOS": "local"},
                cwd="/fixture", exit=0, command=["/bin/sidefx", "-view", "effects", "-json",
                "-max-sites", "10000", *flags, "./..."])


class OperationIndexTests(unittest.TestCase):
    def test_distinct_membership_origin_possible_uncertainty_and_unresolved(self):
        r = SavedReport(fixture(), "hash")
        index = r.index()
        rows = {row["category"]: row for row in index["categories"]}
        self.assertEqual(rows["a"], dict(category="a", sites=2, established=2, possible_only=0,
                                       uncertain=1, unresolved=0, selected_origin=1, dependency_origin=1, unknown_origin=0))
        self.assertEqual(rows["c"]["possible_only"], 2)
        self.assertEqual(rows["c"]["unknown_origin"], 1)
        self.assertEqual(rows["c"]["uncertain"], 1)
        self.assertEqual(rows["c"]["unresolved"], 1)
        self.assertEqual(index["category_site_memberships"], 5)
        self.assertEqual(index["categorized_distinct_sites"], 4)
        self.assertEqual(index["multi_category_sites"], 1)
        self.assertEqual(index["bounds"]["uncategorized_unresolved_records"], 1)
        self.assertEqual(index["bounds"]["possible_only_unresolved_overlap"], 1)
        self.assertEqual(index["bounds"]["registry"].split(";")[0], "unavailable")
        self.assertEqual(index["bounds"]["declared_query_categories"], "not_established")

    def test_pages_are_lossless_and_preserve_contracts(self):
        data = fixture()
        r = SavedReport(data, "hash")
        first = r.page(category="a", limit=1)
        second = r.page(category="a", offset=first["next_offset"], limit=1)
        self.assertEqual(first["sites"] + second["sites"], data["sites"][:2])
        self.assertEqual(first["interface_contracts"], data["interface_contracts"])
        self.assertEqual(first["reverse"], data["reverse"][:1])
        self.assertIsNone(second["next_offset"])
        self.assertEqual(r.page(uncategorized=True)["sites"], data["sites"][4:])
        self.assertEqual(r.page(site="site:3")["sites"], [data["sites"][3]])
        for kwargs in ({"category": "missing"}, {"site": "missing"}, {"category": "a", "offset": 3},
                       {"category": "a", "limit": 101}, {"category": "a", "offset": -1}):
            with self.subTest(kwargs=kwargs), self.assertRaises(ValueError):
                r.page(**kwargs)

    def test_reverse_edge_endpoints_must_belong_to_each_reverse_graph(self):
        for missing in ("caller", "callee"):
            data = fixture()
            edge = dict(id="edge:0", caller="node:0", callee="node:0",
                        synthetic=True, candidate=True)
            edge[missing] = "node:1"
            data["edges"] = [edge]
            data["reverse"][0]["edges"] = ["edge:0"]
            with self.subTest(missing=missing), self.assertRaisesRegex(ValueError, "reverse edge endpoint"):
                SavedReport(data, "hash")

    def test_pages_preserve_closed_reverse_references(self):
        data = fixture()
        data["edges"] = [dict(id="edge:0", caller="node:1", callee="node:0",
                              synthetic=True, candidate=True)]
        data["reverse"][0]["edges"] = ["edge:0"]
        data["reverse"][0]["nodes"].append({"id": "node:1", "depth": 1})
        saved = SavedReport(data, "hash")
        for selector in ({"site": "site:0"}, {"category": "a", "limit": 1}):
            page = saved.page(**selector)
            nodes = {node["id"] for node in page["nodes"]}
            edges = {edge["id"]: edge for edge in page["edges"]}
            self.assertEqual(page["edges"], data["edges"])
            for reverse in page["reverse"]:
                refs = {node["id"] for node in reverse["nodes"]}
                self.assertLessEqual(refs, nodes)
                for eid in reverse["edges"]:
                    self.assertIn(edges[eid]["caller"], refs)
                    self.assertIn(edges[eid]["callee"], refs)

    def test_filtered_acquisition_is_retained_without_claiming_completeness(self):
        supplied = provenance("hash", "-effect", "a", "-max-depth", "0", "-max-edges", "9")
        supplied["build"]["source_identity"] = "recorded dirty fixture identity"
        for flags in (supplied, provenance("hash", "-unsupported-flag", "value")):
            bounds = SavedReport(fixture(), "hash").index(flags)["bounds"]
            self.assertEqual(bounds["declared_query_categories"], "not_established")
            self.assertEqual(bounds["acquisition"], flags)
            self.assertIsNone(bounds["query"])
            self.assertIn("filtered/unsupported", bounds["orientation_limits"][0])
            self.assertIn("acquisition.command", bounds["source_filters"])
            self.assertIn("acquisition.command", bounds["output_limits"])
        qualified = SavedReport(fixture(), "hash").index(provenance())["bounds"]
        self.assertEqual(qualified["acquisition"], provenance())
        self.assertNotIn("provenance", qualified["query"])

    def test_filtered_provenance_is_hash_guarded_before_eligibility(self):
        # Acquisition/exclusive collision behavior now has native installed-CLI
        # tests. The development reference only projects saved evidence.
        report = SavedReport(fixture(), "hash")
        with self.assertRaisesRegex(ValueError, "SHA-256 mismatch"):
            report.index(provenance("stale", "-effect", "a"))
        record = provenance("hash", "-effect", "a")
        bounds = report.index(record)["bounds"]
        self.assertEqual(bounds["acquisition"], record)
        self.assertEqual(bounds["declared_query_categories"], "not_established")

    def test_declared_scope_and_output_loss_are_independent_from_callers(self):
        data = fixture()
        r = SavedReport(data, "hash")
        bounds = r.index(provenance("hash", "-direct"))["bounds"]
        self.assertEqual(bounds["declared_query_categories"], "complete")
        self.assertTrue(bounds["truncated"]["depth"])
        self.assertEqual(bounds["query"]["scope"], "direct selected roots")
        data["coverage"]["sites_matching_filters"] += 1
        data["truncated"]["sites"] = True
        self.assertEqual(SavedReport(data, "hash").index(provenance())["bounds"]
                         ["declared_query_categories"], "not_established")
        for flags in (("-effect", "a"), ("-function", "F"), ("-package", "p"), ("-unresolved",),
                      ("-api", "API"), ("-reason", "callback"), ("-site", "site:0")):
            query, reason = declared_query(provenance("hash", *flags), "hash")
            self.assertIsNone(query)
            self.assertIn("unsupported", reason[0])
        bad = provenance("wrong")
        with self.assertRaises(ValueError):
            r.index(bad)
        for field, value in (("producer", {}), ("build", True), ("cwd", False)):
            bad = provenance()
            bad[field] = value
            with self.subTest(field=field), self.assertRaises(ValueError):
                r.index(bad)

    def test_rejects_malformed_or_unsupported_inputs(self):
        cases = [[], {"schema_version": 2}, fixture(), fixture(), fixture(), fixture(), fixture()]
        cases[2]["sites"].append(copy.deepcopy(cases[2]["sites"][0]))
        cases[3]["coverage"]["sites_emitted"] = 1
        cases[4]["reverse"][0]["origins"] = ["node:missing"]
        cases[5]["groups"] = [{"effect": "a", "sites": ["site:2"]}]
        cases[6]["sites"][0]["effects"] = ["a", "a"]
        for data in cases:
            with self.subTest(data=data), self.assertRaises(ValueError):
                SavedReport(data, "hash")

    def test_cli_stable_bytes_input_identity_and_no_partial_failure(self):
        with tempfile.TemporaryDirectory() as root:
            path = Path(root) / "effects.json"
            path.write_text(json.dumps(fixture()))
            argv = [sys.executable, str(Path(__file__).with_name("operation_index.py")), str(path)]
            first = subprocess.run(argv, capture_output=True, check=True)
            second = subprocess.run(argv, capture_output=True, check=True)
            self.assertEqual(first.stdout, second.stdout)
            self.assertEqual(json.loads(first.stdout)["input_sha256"], hashlib.sha256(path.read_bytes()).hexdigest())
            bad = subprocess.run(argv + ["--expect-sha256", "wrong"], capture_output=True)
            self.assertEqual(bad.returncode, 1)
            self.assertEqual(bad.stdout, b"")
            self.assertIn(b"SHA-256 mismatch", bad.stderr)


if __name__ == "__main__":
    unittest.main()
