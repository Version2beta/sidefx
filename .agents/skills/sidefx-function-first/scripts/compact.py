#!/usr/bin/env python3
"""Reduce sidefx function JSON without promoting candidate evidence."""

import json
import sys


def compact(records):
    if not isinstance(records, list):
        raise ValueError("expected a function-view JSON array")
    rows = []
    for index, record in enumerate(records):
        if not isinstance(record, dict):
            raise ValueError(f"record {index}: expected an object")
        for key in ("package", "function"):
            if not isinstance(record.get(key), str) or not record[key]:
                raise ValueError(f"record {index}: missing/invalid {key}")
        if type(record.get("uncertain")) is not bool:
            raise ValueError(f"record {index}: missing/invalid uncertain boolean")
        effects = record.get("effects")
        if not isinstance(effects, list):
            raise ValueError(f"record {index}: missing/invalid effects array")
        direct, inherited = [], []
        for effect in effects:
            if (not isinstance(effect, dict)
                    or not isinstance(effect.get("identifier"), str)
                    or not effect["identifier"]
                    or type(effect.get("direct")) is not bool):
                raise ValueError(f"record {index}: invalid established effect")
            (direct if effect["direct"] else inherited).append(effect["identifier"])
        possible = record.get("possible_effects", [])
        if (not isinstance(possible, list)
                or any(not isinstance(item, str) or not item for item in possible)):
            raise ValueError(f"record {index}: invalid possible_effects array")
        rows.append({
            "package": record["package"],
            "function": record["function"],
            "location": None,
            "direct_marked": direct,
            "inherited_marked": inherited,
            "possible_effects": possible,
            "uncertain": record["uncertain"],
            "uncertainty_reasons": None,
            "uncertainty_provenance": None,
        })
    return rows


def main():
    if len(sys.argv) != 2:
        print("usage: compact.py FUNCTIONS.json", file=sys.stderr)
        return 2
    try:
        with open(sys.argv[1], encoding="utf-8") as source:
            rows = compact(json.load(source))
        # Validate the full input first: malformed records must not emit partial rows.
        for row in rows:
            print(json.dumps(row, separators=(",", ":"), ensure_ascii=False))
    except (OSError, ValueError) as error:
        print(f"compact: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
