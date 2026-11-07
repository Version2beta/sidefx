#!/usr/bin/env python3
"""Development/reference-only saved-JSON projection for native parity checks.

Normal installed workflows use sidefx acquire / sidefx index, not Python.
No analysis or recognition occurs here.
"""

import argparse
import hashlib
import json
from pathlib import Path
import sys


def require(condition, message):
    if not condition:
        raise ValueError(message)


def strings(value, label):
    require(isinstance(value, list) and all(isinstance(x, str) and x for x in value),
            f"{label}: expected nonempty strings in an array")
    require(len(set(value)) == len(value), f"{label}: duplicate values")
    return set(value)


def objects(value, key, label):
    require(isinstance(value, list), f"{label}: expected array")
    result = {}
    for item in value:
        require(isinstance(item, dict) and isinstance(item.get(key), str) and item[key],
                f"{label}: invalid {key}")
        require(item[key] not in result, f"{label}: duplicate {key} {item[key]}")
        result[item[key]] = item
    return result


class SavedReport:
    def __init__(self, data, digest):
        require(isinstance(data, dict) and type(data.get("schema_version")) is int
                and data["schema_version"] == 1, "expected effect-view schema_version 1")
        self.data, self.digest = data, digest
        self.sites = objects(data.get("sites"), "id", "sites")
        self.nodes = objects(data.get("nodes"), "id", "nodes")
        self.edges = objects(data.get("edges"), "id", "edges")
        self.reverse = objects(data.get("reverse"), "site", "reverse")
        require(set(self.reverse) == set(self.sites), "reverse/site sets differ")
        self.categories = {}
        for sid, site in self.sites.items():
            for key in ("package", "function", "kind", "recognition"):
                require(isinstance(site.get(key), str) and site[key], f"{sid}: invalid {key}")
            require(type(site.get("uncertain")) is bool, f"{sid}: invalid uncertain")
            strings(site.get("apis"), f"{sid}.apis")
            strings(site.get("uncertainty_reasons"), f"{sid}.uncertainty_reasons")
            location = site.get("location")
            require(isinstance(location, dict) and isinstance(location.get("file"), str)
                    and all(type(location.get(k)) is int and location[k] >= 0
                            for k in ("line", "column")), f"{sid}: invalid location")
            require(isinstance(site.get("operations"), list)
                    and all(isinstance(op, dict) for op in site["operations"]),
                    f"{sid}: invalid operations")
            established = strings(site.get("effects"), f"{sid}.effects")
            possible = strings(site.get("possible_effects", []), f"{sid}.possible_effects")
            require(not established & possible, f"{sid}: established/possible overlap")
            for category in established | possible:
                row = self.categories.setdefault(category, {"established": set(), "possible_only": set()})
                row["established" if category in established else "possible_only"].add(sid)
        for node in self.nodes.values():
            require(type(node.get("selected")) is bool, "node: invalid selected")
        for edge in self.edges.values():
            require(edge.get("caller") in self.nodes and edge.get("callee") in self.nodes,
                    "edge: dangling node")
        for sid, reverse in self.reverse.items():
            origins = strings(reverse.get("origins"), f"{sid}.origins")
            refs = objects(reverse.get("nodes"), "id", f"{sid}.nodes")
            require(origins <= set(refs) <= set(self.nodes), f"{sid}: dangling reverse nodes")
            edge_ids = strings(reverse.get("edges"), f"{sid}.edges")
            require(edge_ids <= set(self.edges), f"{sid}: dangling reverse edges")
            for eid in edge_ids:
                edge = self.edges[eid]
                require(edge["caller"] in refs and edge["callee"] in refs,
                        f"{sid}: reverse edge endpoint missing from node refs: {eid}")
            self.check_truncation(reverse.get("truncated"))
        groups = objects(data.get("groups"), "effect", "groups")
        for category, group in groups.items():
            for key, field in (("sites", "established"), ("possible_sites", "possible_only")):
                ids = strings(group.get(key, []), f"group {category}.{key}")
                require(ids <= self.categories.get(category, {}).get(field, set()),
                        f"group {category}: invalid {key} reference")
        self.unresolved = strings(data.get("unresolved_sites"), "unresolved_sites")
        require(self.unresolved == {sid for sid, s in self.sites.items()
                                    if not s["effects"] and s["uncertain"]},
                "unresolved_sites differs from schema's effect-free uncertain sites")
        self.uncategorized = {sid for sid in self.unresolved
                              if not self.sites[sid].get("possible_effects")}
        self.check_truncation(data.get("truncated"))
        contracts = objects(data.get("interface_contracts", []), "site", "interface_contracts")
        require(set(contracts) <= set(self.sites), "contract: dangling site")
        coverage = data.get("coverage")
        require(isinstance(coverage, dict), "missing coverage")
        strings(coverage.get("packages"), "coverage.packages")
        for key in ("sites_emitted", "sites_matching_filters", "operation_sites_identified",
                    "selected_operation_sites", "dependency_operation_sites", "unresolved_sites_identified"):
            require(type(coverage.get(key)) is int and coverage[key] >= 0, f"invalid coverage.{key}")
        require(coverage["sites_emitted"] == len(self.sites), "coverage sites_emitted mismatch")
        require(coverage["sites_matching_filters"] >= len(self.sites), "invalid matching site count")
        require(data["truncated"]["sites"] == (coverage["sites_matching_filters"] > len(self.sites)),
                "site truncation/count mismatch")

    @staticmethod
    def check_truncation(value):
        require(isinstance(value, dict) and all(type(value.get(k)) is bool
                for k in ("sites", "edges", "depth")), "invalid graph truncation")

    def origin(self, sid):
        site = self.sites[sid]
        nodes = [self.nodes[n] for n in self.reverse[sid]["origins"]]
        matching = [n for n in nodes if (n.get("package"), n.get("function"))
                    == (site["package"], site["function"])]
        if not matching:
            return "unknown_origin"
        return "selected_origin" if any(n["selected"] for n in matching) else "dependency_origin"

    def bounds(self, provenance=None):
        query, reasons = declared_query(provenance, self.digest)
        no_site_loss = not self.data["truncated"]["sites"]
        return {
            "emitted_records": "complete",
            "declared_query_categories": "complete" if query and no_site_loss else "not_established",
            "orientation_limits": reasons + ([] if no_site_loss else ["input site output is truncated"]),
            "query": query,
            "acquisition": provenance,
            "analysis_coverage": "recognized vocabulary/current build only; not runtime or all mutations",
            "dependency_analysis_bounds": self.data["coverage"].get("dependency_vta", "not exposed; default discovery remains selected-source bounded"),
            "excluded_scope": "other builds, unselected roots and unmodeled operations; direct scope also excludes dependency-origin sites and candidate-only helper reachability",
            "registry": "unavailable; absent labels are not recognized-zero categories",
            "source_filters": "see acquisition.command; completeness eligibility in orientation_limits" if provenance is not None
                              else "unknown; coverage alone cannot disclose filters",
            "output_limits": "see acquisition.command; unspecified producer defaults unavailable" if provenance is not None
                             else "unavailable; truncation flags do not state numeric limits",
            "truncated": self.data["truncated"],
            "coverage": self.data["coverage"],
            "uncategorized_unresolved_records": len(self.uncategorized),
            "unresolved_view_records": len(self.unresolved),
            "uncategorized_other_records": sum(not s["effects"] and not s.get("possible_effects")
                                                and not s["uncertain"] for s in self.sites.values()),
            "possible_only_unresolved_overlap": len(self.unresolved - self.uncategorized),
            "unresolved_inventory": "not established; identified count is not an emitted inventory",
        }

    def index(self, provenance=None):
        rows = []
        memberships = 0
        categorized = set()
        for category, groups in sorted(self.categories.items()):
            ids = groups["established"] | groups["possible_only"]
            categorized |= ids
            memberships += len(ids)
            counts = {key: sum(self.origin(sid) == key for sid in ids)
                      for key in ("selected_origin", "dependency_origin", "unknown_origin")}
            rows.append({"category": category, "sites": len(ids),
                         "established": len(groups["established"]),
                         "possible_only": len(groups["possible_only"]),
                         "uncertain": sum(self.sites[sid]["uncertain"] for sid in ids),
                         "unresolved": len(ids & self.unresolved), **counts})
        return {"projection": "sidefx-operation-index-v1", "input_sha256": self.digest,
                "bounds": self.bounds(provenance), "categories": rows,
                "category_count": len(rows), "categorized_distinct_sites": len(categorized),
                "category_site_memberships": memberships,
                "multi_category_sites": sum(len(s["effects"]) + len(s.get("possible_effects", [])) > 1
                                            for s in self.sites.values()),
                "count_definitions": "Distinct canonical site IDs per category; established/possible_only are origin labels, not caller-path certainty. possible_only is per category. uncertain overlaps both. Origin counts partition sites by matching reverse origin nodes; not direct/inherited function marks. Category counts overlap across multi-effect sites; do not sum as distinct sites."}

    def page(self, category=None, site=None, uncategorized=False, offset=0, limit=10, provenance=None):
        if category is not None:
            require(category in self.categories, f"category not present: {category}")
            groups = self.categories[category]
            ids = groups["established"] | groups["possible_only"]
        elif site is not None:
            require(site in self.sites, f"site not present: {site}")
            ids = {site}
        else:
            require(uncategorized, "retrieval requires category, site or uncategorized unresolved selector")
            ids = self.uncategorized
        require(0 <= offset <= len(ids), "offset outside selected site set")
        require(1 <= limit <= 100, "limit must be 1..100")
        selected = sorted(ids)[offset:offset + limit]
        reverse = [self.reverse[sid] for sid in selected]
        nodes = {n["id"] for r in reverse for n in r["nodes"]}
        edges = {e for r in reverse for e in r["edges"]}
        return {"projection": "sidefx-operation-sites-v1", "input_sha256": self.digest,
                "selector": {"category": category, "site": site, "uncategorized_unresolved": uncategorized},
                "total": len(ids), "offset": offset, "limit": limit,
                "next_offset": offset + len(selected) if offset + len(selected) < len(ids) else None,
                "bounds": self.bounds(provenance),
                "sites": [self.sites[sid] for sid in selected], "reverse": reverse,
                "nodes": [self.nodes[n] for n in sorted(nodes)],
                "edges": [self.edges[e] for e in sorted(edges)],
                "interface_contracts": [c for c in self.data.get("interface_contracts", [])
                                        if c.get("site") in selected]}


def declared_query(provenance, digest):
    """Attest a recorded, unfiltered CLI acquisition, not authenticity or coverage."""
    if provenance is None:
        return None, ["producer command/build/scope provenance unavailable"]
    require(isinstance(provenance, dict), "provenance: expected object")
    require(provenance.get("input_sha256") == digest, "provenance: input SHA-256 mismatch")
    for key in ("producer", "build"):
        require(isinstance(provenance.get(key), dict) and provenance[key],
                f"provenance: missing/invalid {key} object")
    producer_hash = provenance["producer"].get("sha256")
    require(isinstance(producer_hash, str) and len(producer_hash) == 64
            and all(c in "0123456789abcdef" for c in producer_hash),
            "provenance: invalid producer SHA-256")
    require(isinstance(provenance.get("cwd"), str) and provenance["cwd"],
            "provenance: missing/invalid cwd")
    require(type(provenance.get("exit")) is int and provenance["exit"] == 0,
            "provenance: acquisition did not succeed")
    argv = provenance.get("command")
    require(isinstance(argv, list) and len(argv) > 1 and all(isinstance(x, str) and x for x in argv),
            "provenance: command must be argv array")
    values, patterns = {}, []
    booleans = {"json", "direct", "tests", "exact", "stats", "all"}
    valued = {"view", "max-sites", "max-edges", "max-depth", "dependency-vta-functions", "dependency-vta-edges"}
    args = iter(argv[1:])
    for arg in args:
        if not arg.startswith("-") or arg == "-":
            patterns = [arg, *args]
            break
        name, equals, value = arg.lstrip("-").partition("=")
        if name not in booleans | valued or name in values:
            return None, [f"filtered/unsupported or duplicate acquisition flag: {name}"]
        if name in booleans:
            require(not equals or value in ("true", "false"), f"provenance: invalid -{name}")
            values[name] = value != "false"
        else:
            value = value if equals else next(args, None)
            require(value is not None, f"provenance: missing -{name} value")
            if name != "view":
                require(value.isdecimal() and int(value) >= (0 if name == "max-depth" else 1),
                        f"provenance: invalid -{name}")
            values[name] = value
    require(patterns and all(not p.startswith("-") for p in patterns), "provenance: explicit package scope required; flags must precede packages")
    require(values.get("view") == "effects" and values.get("json") is True,
            "provenance: require -view effects -json")
    require("max-sites" in values, "provenance: record explicit positive -max-sites")
    require(("dependency-vta-functions" in values) == ("dependency-vta-edges" in values),
            "provenance: dependency VTA budgets must be paired")
    return {"scope": "direct selected roots" if values.get("direct") else "reachable from selected roots (including candidate paths)",
            "patterns": patterns, "flags": values,
            "authority": "caller-supplied acquisition record, hash-bound but not authenticated"}, []


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", type=Path)
    parser.add_argument("--provenance", type=Path, help="hash-bound recorded acquisition JSON; see docs/operation-index.md")
    parser.add_argument("--expect-sha256", help="reject a changed input between pages")
    select = parser.add_mutually_exclusive_group()
    select.add_argument("--category", help="retrieve exact category, including possible-only sites")
    select.add_argument("--site", help="retrieve exact canonical site ID")
    select.add_argument("--uncategorized-unresolved", action="store_true")
    parser.add_argument("--offset", type=int, default=0)
    parser.add_argument("--limit", type=int, default=10, help="sites per page (1..100); graphs retain their input bounds")
    args = parser.parse_args(argv)
    try:
        raw = args.report.read_bytes()
        digest = hashlib.sha256(raw).hexdigest()
        require(args.expect_sha256 is None or args.expect_sha256 == digest, "input SHA-256 mismatch")
        saved = SavedReport(json.loads(raw), digest)
        provenance = json.loads(args.provenance.read_bytes()) if args.provenance else None
        if args.category is not None or args.site is not None or args.uncategorized_unresolved:
            result = saved.page(args.category, args.site, args.uncategorized_unresolved,
                                args.offset, args.limit, provenance)
        else:
            require(args.offset == 0 and args.limit == 10, "offset/limit require a retrieval selector")
            result = saved.index(provenance)
        print(json.dumps(result, sort_keys=True, separators=(",", ":"), ensure_ascii=False))
    except (OSError, ValueError) as error:
        print(f"operation_index: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
